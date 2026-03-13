package queue

import (
	"os"
)

type RetentionPolicy struct {
	MaxSize int64
}

func (p *Partition) Retention(policy RetentionPolicy) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if policy.MaxSize <= 0 {
		return nil
	}

	var totalSize int64
	keepIdx := 0
	
	// Collect total size from most recent to oldest
	for i := len(p.segments) - 1; i >= 0; i-- {
		totalSize += p.segments[i].Size()
		if totalSize > policy.MaxSize {
			keepIdx = i + 1
			break
		}
	}

	if keepIdx > 0 {
		toDelete := p.segments[:keepIdx]
		p.segments = p.segments[keepIdx:]

		for _, s := range toDelete {
			s.Close()
			os.Remove(s.logFile.Name())
			os.Remove(s.indexFile.Name())
		}
	}

	return nil
}
