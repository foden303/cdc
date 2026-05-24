package service

import (
	"context"

	"github.com/foden/cdc/pkg/dto/request"
	"github.com/foden/cdc/pkg/dto/response"
	"github.com/foden/cdc/pkg/interfaces"
)

type DLQService struct {
	natsClient interfaces.NATSClient
}

func NewDLQService(natsClient interfaces.NATSClient) *DLQService {
	return &DLQService{natsClient: natsClient}
}

func (s *DLQService) ListMessages(ctx context.Context, req request.ListDLQMessagesRequest) (response.ListDLQMessagesResponse, error) {
	if s.natsClient == nil {
		return response.ListDLQMessagesResponse{Pagination: pagination(0, req.Page, req.Limit)}, nil
	}
	items, total, err := s.natsClient.ListDLQMessages(ctx, normalizedLimit(req.Limit), normalizedPage(req.Page))
	if err != nil {
		return response.ListDLQMessagesResponse{}, err
	}
	result := make([]response.DLQMessage, 0, len(items))
	for _, item := range items {
		result = append(result, response.DLQMessage{
			Message:         item,
			Reason:          item.Headers["X-DLQ-Reason"],
			OriginalSubject: item.Headers["X-DLQ-Original-Subject"],
		})
	}
	return response.ListDLQMessagesResponse{
		Data:       result,
		Pagination: pagination(total, req.Page, req.Limit),
	}, nil
}

func (s *DLQService) Reprocess(ctx context.Context, _ request.ReprocessDLQRequest) (response.ReprocessDLQResponse, error) {
	if s.natsClient == nil {
		return response.ReprocessDLQResponse{Count: 0}, nil
	}
	count, err := s.natsClient.ReprocessDLQ(ctx)
	if err != nil {
		return response.ReprocessDLQResponse{}, err
	}
	return response.ReprocessDLQResponse{Count: int32(count)}, nil
}
