package response

import "github.com/foden/cdc/internal/core/ports"

type ReprocessDLQResponse struct {
	Count int32
}

type DLQMessage struct {
	Message         *ports.NATSMessageItem
	Reason          string
	OriginalSubject string
}

type ListDLQMessagesResponse struct {
	Data       []DLQMessage
	Pagination PaginationResponse
}
