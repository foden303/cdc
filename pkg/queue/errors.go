package queue

import "errors"

var (
	ErrSegmentFull = errors.New("segment full")
	ErrCRC         = errors.New("crc mismatch")
	ErrQueueClosed = errors.New("queue closed")
	ErrClosed      = errors.New("segment closed")
)
