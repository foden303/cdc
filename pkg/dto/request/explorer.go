package request

import "github.com/foden/cdc/pkg/models"

type ListMessagesRequest struct {
	Status    models.MessageStatus
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
