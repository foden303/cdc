package common

import (
	"github.com/foden/cdc/internal/core/constant"
	"github.com/foden/cdc/internal/core/domain"
	"github.com/foden/cdc/pkg/pool"
)

// BuildEvent creates a pooled CDC envelope with normalized shared fields.
// All sources should use this builder to keep event metadata consistent.
func BuildEvent(
	topic string,
	subject string,
	instanceID string,
	schema string,
	table string,
	op constant.Op,
	lsn uint64,
	offset string,
	data []byte,
	partition int,
	messageID string,
) *domain.Event {
	ev := pool.GetEvent()
	ev.Topic = topic
	ev.Subject = subject
	ev.InstanceID = instanceID
	ev.Schema = schema
	ev.Table = table
	ev.Op = op
	ev.LSN = lsn
	ev.Offset = offset
	ev.Data = data
	ev.Partition = partition
	ev.MessageID = messageID
	return ev
}
