package queue

// Message is a message in the queue. Model Write
type Message struct {
	Offset    uint64
	Key       []byte
	Value     []byte
	Timestamp int64
	Retry     int
}

// MessageView is a message in the queue. Model Read
type MessageView struct {
	Offset    uint64
	Key       []byte
	Value     []byte
	Timestamp int64
	Retry     int
}

// QueueStats holds partition-level statistics.
type QueueStats struct {
	SegmentsCount int
	TotalSizeMB   int64
	TotalEnqueued uint64
	TotalDequeued uint64
	Pending       uint64
}
