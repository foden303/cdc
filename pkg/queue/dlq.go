package queue

type DLQ struct {
	topic *Topic
}

func (d *DLQ) Send(msg *MessageView) {
	// Simple DLQ implementation just produces to the designated topic
	retryMsg := &Message{
		Key:       msg.Key,
		Value:     msg.Value,
		Timestamp: msg.Timestamp,
		Retry:     msg.Retry,
	}
	d.topic.Partition(msg.Key).Produce(retryMsg)
}
