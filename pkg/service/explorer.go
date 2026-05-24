package service

import (
	"context"
	"strings"

	"github.com/foden/cdc/pkg/dto/request"
	"github.com/foden/cdc/pkg/dto/response"
	"github.com/foden/cdc/pkg/interfaces"
)

type ExplorerService struct {
	natsClient interfaces.NATSClient
}

func NewExplorerService(natsClient interfaces.NATSClient) *ExplorerService {
	return &ExplorerService{natsClient: natsClient}
}

func (s *ExplorerService) Messages(ctx context.Context, req request.ListMessagesRequest) (response.ListMessagesResponse, error) {
	if s.natsClient == nil {
		return response.ListMessagesResponse{Pagination: pagination(0, req.Page, req.Limit)}, nil
	}
	items, total, err := s.natsClient.ListMessages(ctx, req.Status, normalizedLimit(req.Limit), normalizedPage(req.Page), req.Topic, req.Partition)
	if err != nil {
		return response.ListMessagesResponse{}, err
	}
	return response.ListMessagesResponse{
		Data:       items,
		TotalCount: total,
		Pagination: pagination(total, req.Page, req.Limit),
	}, nil
}

func (s *ExplorerService) Topics(ctx context.Context, req request.ListTopicsRequest) (response.ListTopicsResponse, error) {
	if s.natsClient == nil {
		return response.ListTopicsResponse{Pagination: pagination(0, req.Page, req.Limit)}, nil
	}
	topics, total, err := s.natsClient.ListTopics(ctx, normalizedLimit(req.Limit), normalizedPage(req.Page))
	if err != nil {
		return response.ListTopicsResponse{}, err
	}
	result := make([]response.TopicSummary, 0, len(topics))
	for _, topic := range topics {
		partitions, _, partitionsErr := s.natsClient.ListPartitions(ctx, topic, 500, 1)
		partitionCount := int32(0)
		if partitionsErr == nil {
			partitionCount = int32(len(partitions))
		}
		result = append(result, response.TopicSummary{
			Name:           topic,
			PartitionCount: partitionCount,
			MessageCount:   0,
		})
	}
	return response.ListTopicsResponse{Data: result, Pagination: pagination(total, req.Page, req.Limit)}, nil
}

func (s *ExplorerService) Partitions(ctx context.Context, req request.ListPartitionsRequest) (response.ListPartitionsResponse, error) {
	if s.natsClient == nil {
		return response.ListPartitionsResponse{Pagination: pagination(0, req.Page, req.Limit)}, nil
	}
	partitions, total, err := s.natsClient.ListPartitions(ctx, req.Topic, normalizedLimit(req.Limit), normalizedPage(req.Page))
	if err != nil {
		return response.ListPartitionsResponse{}, err
	}
	result := make([]response.PartitionSummary, 0, len(partitions))
	for _, subject := range partitions {
		result = append(result, response.PartitionSummary{
			ID:    subject,
			Topic: topicFromSubject(subject),
		})
	}
	return response.ListPartitionsResponse{Data: result, Pagination: pagination(total, req.Page, req.Limit)}, nil
}

func (s *ExplorerService) Consumers(ctx context.Context, req request.ListConsumersRequest) (response.ListConsumersResponse, error) {
	if s.natsClient == nil {
		return response.ListConsumersResponse{Pagination: pagination(0, req.Page, req.Limit)}, nil
	}
	consumers, total, err := s.natsClient.ListConsumers(ctx, normalizedLimit(req.Limit), normalizedPage(req.Page))
	if err != nil {
		return response.ListConsumersResponse{}, err
	}
	return response.ListConsumersResponse{Data: consumers, Pagination: pagination(total, req.Page, req.Limit)}, nil
}

func topicFromSubject(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) < 4 {
		return subject
	}
	return strings.Join(parts[:4], ".")
}

func normalizedLimit(limit int) int {
	if limit <= 0 {
		return 25
	}
	if limit > 500 {
		return 500
	}
	return limit
}

func normalizedPage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func pagination(total uint64, page int, limit int) response.PaginationResponse {
	page = normalizedPage(page)
	limit = normalizedLimit(limit)
	start := uint64((page - 1) * limit)
	return response.PaginationResponse{
		TotalRows: total,
		Page:      int32(page),
		Limit:     int32(limit),
		HasPrev:   page > 1,
		HasNext:   start+uint64(limit) < total,
	}
}
