package response

import "github.com/foden/cdc/pkg/interfaces"

type PaginationResponse struct {
	TotalRows uint64
	Page      int32
	Limit     int32
	HasNext   bool
	HasPrev   bool
}

type TopicSummary struct {
	Name           string
	MessageCount   uint64
	PartitionCount int32
}

type PartitionSummary struct {
	ID           string
	MessageCount uint64
	Topic        string
}

type ListMessagesResponse struct {
	Data       []*interfaces.NATSMessageItem
	TotalCount uint64
	Pagination PaginationResponse
}

type ListTopicsResponse struct {
	Data       []TopicSummary
	Pagination PaginationResponse
}

type ListPartitionsResponse struct {
	Data       []PartitionSummary
	Pagination PaginationResponse
}

type ListConsumersResponse struct {
	Data       []interfaces.NATSConsumerSummary
	Pagination PaginationResponse
}
