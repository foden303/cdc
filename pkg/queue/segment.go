package queue

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
)

type IndexEntry struct {
	Offset uint32
	Pos    uint32
}

type Segment struct {
	baseOffset uint64
	maxSize    int64

	logFile   *os.File
	indexFile *os.File

	mmap []byte

	writePos   atomic.Int64
	visiblePos atomic.Int64
	lastOffset atomic.Uint64

	mu sync.Mutex

	bytesSinceLastIndex int64
	indexCache          []IndexEntry // In-memory index cache
	indexInterval       int64        // Configurable index interval
}

const indexInterval = 4096 // 4KB

func OpenSegment(path string, baseOffset uint64, maxSize int64, indexInterval int64) (*Segment, error) {
	if indexInterval <= 0 {
		indexInterval = 4096
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	info, _ := f.Stat()
	size := info.Size()

	if size == 0 {
		if err := f.Truncate(maxSize); err != nil {
			return nil, err
		}
	}

	m, err := syscall.Mmap(int(f.Fd()), 0, int(maxSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	idxPath := path[:len(path)-len(filepath.Ext(path))] + ".index"
	idxF, err := os.OpenFile(idxPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	s := &Segment{
		baseOffset:    baseOffset,
		maxSize:       maxSize,
		logFile:       f,
		indexFile:     idxF,
		mmap:          m,
		indexInterval: indexInterval,
	}

	if size > 0 {
		s.recover(size)
	}

	// Load index cache
	if indexCache, err := s.loadIndex(); err == nil {
		s.indexCache = indexCache
	}

	return s, nil
}

func (s *Segment) recover(fileSize int64) {
	var pos int64
	var lastOffset uint64

	for pos < fileSize {
		if pos+4 > fileSize {
			break
		}
		sizeField := int64(binary.LittleEndian.Uint32(s.mmap[pos : pos+4]))
		if sizeField == 0 {
			break
		}

		recordSize := 4 + sizeField // size_field(4) + CRC+payload(sizeField)
		if pos+recordSize > fileSize {
			break
		}

		// Verify CRC before accepting record
		crcStored := binary.LittleEndian.Uint32(s.mmap[pos+4 : pos+8])
		payload := s.mmap[pos+8 : pos+recordSize]
		if crc32.ChecksumIEEE(payload) != crcStored {
			break
		}

		offset := binary.LittleEndian.Uint64(payload[0:8])
		lastOffset = offset
		pos += recordSize
	}

	s.writePos.Store(pos)
	s.visiblePos.Store(pos)
	s.lastOffset.Store(lastOffset)
}

func (s *Segment) ReadAt(pos int64) (*MessageView, int64, error) {
	if pos >= s.visiblePos.Load() {
		return nil, 0, os.ErrNotExist
	}

	size := binary.LittleEndian.Uint32(s.mmap[pos : pos+4])
	buf := s.mmap[pos : pos+int64(size)+4]

	crc := binary.LittleEndian.Uint32(buf[4:8])
	payload := buf[8:]
	if crc32.ChecksumIEEE(payload) != crc {
		return nil, 0, errors.New("crc mismatch")
	}

	offset := binary.LittleEndian.Uint64(payload[0:8])
	timestamp := int64(binary.LittleEndian.Uint64(payload[8:16]))
	keyLen := binary.LittleEndian.Uint32(payload[16:20])
	key := payload[20 : 20+keyLen]

	vpos := 20 + keyLen
	valLen := binary.LittleEndian.Uint32(payload[vpos : vpos+4])
	value := payload[vpos+4 : vpos+4+valLen]

	msg := &MessageView{
		Offset:    offset,
		Key:       key,
		Value:     value,
		Timestamp: timestamp,
	}

	return msg, pos + int64(size) + 4, nil
}

func (s *Segment) Append(msg *Message) (uint64, int64, error) {
	offsets, err := s.AppendBatch([]*Message{msg})
	if err != nil {
		return 0, 0, err
	}
	return offsets[0], 0, nil // pos isn't strictly needed here for Produce
}

func (s *Segment) FetchBatch(pos int64, maxBytes int) ([]*MessageView, int64, error) {
	var res []*MessageView
	read := 0

	for {
		msg, next, err := s.ReadAt(pos)
		if err != nil {
			if err == io.EOF && len(res) > 0 {
				return res, pos, nil
			}
			return res, pos, err
		}

		size := int(next - pos)
		if read+size > maxBytes && len(res) > 0 {
			return res, pos, nil
		}

		res = append(res, msg)
		read += size
		pos = next
	}
}

func (s *Segment) loadIndex() ([]IndexEntry, error) {
	info, err := s.indexFile.Stat()
	if err != nil {
		return nil, err
	}

	count := info.Size() / 8
	entries := make([]IndexEntry, count)

	data := make([]byte, info.Size())
	_, err = s.indexFile.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}

	for i := 0; i < int(count); i++ {
		entries[i].Offset = binary.LittleEndian.Uint32(data[i*8 : i*8+4])
		entries[i].Pos = binary.LittleEndian.Uint32(data[i*8+4 : i*8+8])
	}

	return entries, nil
}
func (s *Segment) FindOffset(offset uint64) (int64, error) {
	// Use cached index instead of reloading from file
	s.mu.Lock()
	entries := s.indexCache
	s.mu.Unlock()

	if len(entries) == 0 {
		return 0, nil
	}

	rel := uint32(offset - s.baseOffset)

	lo := 0
	hi := len(entries) - 1

	for lo <= hi {
		mid := (lo + hi) / 2
		if entries[mid].Offset == rel {
			return int64(entries[mid].Pos), nil
		}
		if entries[mid].Offset < rel {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	if hi < 0 {
		return 0, nil
	}

	return int64(entries[hi].Pos), nil
}

func (s *Segment) AppendBatch(msgs []*Message) ([]uint64, error) {

	var total int64

	for _, m := range msgs {

		keyLen := len(m.Key)
		valLen := len(m.Value)

		payload := 8 + 8 + 4 + keyLen + 4 + valLen
		total += int64(4 + 4 + payload)
	}

	for {

		start := s.writePos.Load()

		if start+total > s.maxSize {
			return nil, ErrSegmentFull
		}

		if s.writePos.CompareAndSwap(start, start+total) {

			pos := start

			offsets := make([]uint64, len(msgs))

			for i, m := range msgs {
				var offset uint64
				if m.Offset > 0 {
					offset = m.Offset
					// Update lastOffset if needed
					for {
						curr := s.lastOffset.Load()
						if offset <= curr || s.lastOffset.CompareAndSwap(curr, offset) {
							break
						}
					}
				} else {
					offset = s.lastOffset.Add(1)
				}
				offsets[i] = offset

				keyLen := len(m.Key)
				valLen := len(m.Value)

				payload := 8 + 8 + 4 + keyLen + 4 + valLen
				recordSize := int64(4 + 4 + payload)

				buf := s.mmap[pos : pos+recordSize]

				binary.LittleEndian.PutUint32(buf[0:4], uint32(payload+4))

				payloadBuf := buf[8:]

				binary.LittleEndian.PutUint64(payloadBuf[0:8], offset)
				binary.LittleEndian.PutUint64(payloadBuf[8:16], uint64(m.Timestamp))

				binary.LittleEndian.PutUint32(payloadBuf[16:20], uint32(keyLen))

				copy(payloadBuf[20:], m.Key)

				kpos := 20 + keyLen

				binary.LittleEndian.PutUint32(payloadBuf[kpos:kpos+4], uint32(valLen))

				copy(payloadBuf[kpos+4:], m.Value)

				crc := crc32.ChecksumIEEE(payloadBuf)

				binary.LittleEndian.PutUint32(buf[4:8], crc)

				pos += recordSize
			}

			s.visiblePos.Store(pos)

			// Sparse indexing
			s.mu.Lock()
			s.bytesSinceLastIndex += total
			if s.bytesSinceLastIndex >= s.indexInterval {
				lastIdx := len(msgs) - 1
				relOffset := uint32(offsets[lastIdx] - s.baseOffset)

				var buf [8]byte
				binary.LittleEndian.PutUint32(buf[0:4], relOffset)
				binary.LittleEndian.PutUint32(buf[4:8], uint32(start))

				s.indexFile.Write(buf[:])

				// Update in-memory cache
				s.indexCache = append(s.indexCache, IndexEntry{Offset: relOffset, Pos: uint32(start)})

				s.bytesSinceLastIndex = 0
			}
			s.mu.Unlock()

			return offsets, nil
		}
	}
}

func (s *Segment) Size() int64 {
	return s.visiblePos.Load()
}

func (s *Segment) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.mmap != nil {
		if err := syscall.Munmap(s.mmap); err != nil {
			slog.Error("failed to unmap segment", "err", err)
		}
		s.mmap = nil
	}
	if err := s.logFile.Close(); err != nil {
		return err
	}
	return s.indexFile.Close()
}
