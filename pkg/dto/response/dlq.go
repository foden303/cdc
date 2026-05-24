package response

import "github.com/foden/cdc/pkg/interfaces"

type ReprocessDLQResponse struct {
	Count int32
}

type DLQMessage struct {
	Message         *interfaces.NATSMessageItem
	Reason          string
	OriginalSubject string
}

type ListDLQMessagesResponse struct {
	Data       []DLQMessage
	Pagination PaginationResponse
}
