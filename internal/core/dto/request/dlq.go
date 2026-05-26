package request

type ReprocessDLQRequest struct{}

type ListDLQMessagesRequest struct {
	Page  int
	Limit int
}
