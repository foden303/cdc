package cluster

import "encoding/json"

// CommandType represents the type of Raft command.
type CommandType uint8

const (
	CmdProduce CommandType = iota
	CmdCreateTopic
)

// Command is a serializable Raft log entry.
type Command struct {
	Type CommandType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ProduceCommand is the payload for a Produce operation.
type ProduceCommand struct {
	Topic     string `json:"topic"`
	Key       []byte `json:"key"`
	Value     []byte `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// CreateTopicCommand is the payload for creating a topic.
type CreateTopicCommand struct {
	Name       string `json:"name"`
	Partitions int    `json:"partitions"`
}

// EncodeCommand serializes a command for the Raft log.
func EncodeCommand(cmdType CommandType, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	cmd := Command{
		Type: cmdType,
		Data: data,
	}

	return json.Marshal(cmd)
}

// ValidateCommand validates a command before applying to Raft.
func ValidateCommand(cmdType CommandType, payload any) error {
	switch cmdType {
	case CmdProduce:
		_, ok := payload.(*ProduceCommand)
		if !ok {
			return &ValidationError{"payload must be ProduceCommand"}
		}
		prod := payload.(*ProduceCommand)
		if prod.Topic == "" {
			return &ValidationError{"topic cannot be empty"}
		}
	case CmdCreateTopic:
		_, ok := payload.(*CreateTopicCommand)
		if !ok {
			return &ValidationError{"payload must be CreateTopicCommand"}
		}
		create := payload.(*CreateTopicCommand)
		if create.Name == "" {
			return &ValidationError{"topic name cannot be empty"}
		}
		if create.Partitions <= 0 {
			return &ValidationError{"partitions must be greater than 0"}
		}
	default:
		return &ValidationError{"unknown command type"}
	}
	return nil
}

// ValidationError represents a command validation failure
type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.msg
}
