package request

import "github.com/foden/cdc/internal/core/domain"

type ListMessagesRequest struct {
	Status    domain.MessageStatus
	Topic     string
	Partition string
	Page      int
	Limit     int
}

type ListTopicsRequest struct {
	Page  int
	Limit int
}

type ListPartitionsRequest struct {
	Topic string
	Page  int
	Limit int
}

type ListConsumersRequest struct {
	Page  int
	Limit int
}
