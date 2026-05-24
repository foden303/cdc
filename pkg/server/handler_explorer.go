package server

import (
	"context"
	"log/slog"
	"strconv"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/dto/request"
	"github.com/foden/cdc/pkg/dto/response"
	"github.com/foden/cdc/pkg/interfaces"
	"github.com/foden/cdc/pkg/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CDCService) ListMessages(ctx context.Context, req *cdcpb.ListMessagesRequest) (*cdcpb.ListMessagesResponse, error) {
	result, err := s.explorerService.Messages(ctx, request.ListMessagesRequest{
		Status:    protoMessageStatus(req.Status),
		Topic:     req.GetTopic(),
		Partition: req.GetPartition(),
		Page:      paginationPage(req.Pagination),
		Limit:     paginationLimit(req.Pagination),
	})
	if err != nil {
		slog.Error("ListMessages failed", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to list messages: %v", err)
	}
	return &cdcpb.ListMessagesResponse{
		Data:       messagesToProto(result.Data),
		TotalCount: result.TotalCount,
		Pagination: paginationToProto(result.Pagination),
	}, nil
}

func (s *CDCService) ListTopics(ctx context.Context, req *cdcpb.ListTopicsRequest) (*cdcpb.ListTopicsResponse, error) {
	result, err := s.explorerService.Topics(ctx, request.ListTopicsRequest{
		Page:  paginationPage(req.Pagination),
		Limit: paginationLimit(req.Pagination),
	})
	if err != nil {
		slog.Error("ListTopics failed", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to list topics: %v", err)
	}
	topics := make([]*cdcpb.TopicSummary, 0, len(result.Data))
	for _, topic := range result.Data {
		topics = append(topics, &cdcpb.TopicSummary{
			Name:           topic.Name,
			MessageCount:   topic.MessageCount,
			PartitionCount: topic.PartitionCount,
		})
	}
	return &cdcpb.ListTopicsResponse{Data: topics, Pagination: paginationToProto(result.Pagination)}, nil
}

func (s *CDCService) ListPartitions(ctx context.Context, req *cdcpb.ListPartitionsRequest) (*cdcpb.ListPartitionsResponse, error) {
	result, err := s.explorerService.Partitions(ctx, request.ListPartitionsRequest{
		Topic: req.Topic,
		Page:  paginationPage(req.Pagination),
		Limit: paginationLimit(req.Pagination),
	})
	if err != nil {
		slog.Error("ListPartitions failed", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to list partitions: %v", err)
	}
	partitions := make([]*cdcpb.PartitionSummary, 0, len(result.Data))
	for _, partition := range result.Data {
		partitions = append(partitions, &cdcpb.PartitionSummary{
			Id:           partition.ID,
			MessageCount: partition.MessageCount,
			Topic:        partition.Topic,
		})
	}
	return &cdcpb.ListPartitionsResponse{Data: partitions, Pagination: paginationToProto(result.Pagination)}, nil
}

func (s *CDCService) GetConsumerInfo(ctx context.Context, req *cdcpb.GetConsumerInfoRequest) (*cdcpb.GetConsumerInfoResponse, error) {
	result, err := s.explorerService.Consumers(ctx, request.ListConsumersRequest{Page: 1, Limit: 500})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get consumer info: %v", err)
	}
	for _, consumer := range result.Data {
		if consumer.Name == req.ConsumerName {
			return &cdcpb.GetConsumerInfoResponse{
				AckFloor:     consumer.AckFloorStreamSeq,
				PendingCount: consumer.NumPending,
			}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "consumer %q not found", req.ConsumerName)
}

func (s *CDCService) ListConsumers(ctx context.Context, req *cdcpb.ListConsumersRequest) (*cdcpb.ListConsumersResponse, error) {
	result, err := s.explorerService.Consumers(ctx, request.ListConsumersRequest{
		Page:  paginationPage(req.Pagination),
		Limit: paginationLimit(req.Pagination),
	})
	if err != nil {
		slog.Error("ListConsumers failed", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to list consumers: %v", err)
	}
	consumers := make([]*cdcpb.ConsumerSummary, 0, len(result.Data))
	for _, consumer := range result.Data {
		consumers = append(consumers, &cdcpb.ConsumerSummary{
			Name:               consumer.Name,
			FilterSubjects:     consumer.FilterSubjects,
			NumPending:         consumer.NumPending,
			NumAckPending:      consumer.NumAckPending,
			DeliveredStreamSeq: consumer.DeliveredStreamSeq,
			AckFloorStreamSeq:  consumer.AckFloorStreamSeq,
		})
	}
	return &cdcpb.ListConsumersResponse{Data: consumers, Pagination: paginationToProto(result.Pagination)}, nil
}

func (s *CDCService) ListDLQMessages(ctx context.Context, req *cdcpb.ListDLQMessagesRequest) (*cdcpb.ListDLQMessagesResponse, error) {
	result, err := s.dlqService.ListMessages(ctx, request.ListDLQMessagesRequest{
		Page:  paginationPage(req.Pagination),
		Limit: paginationLimit(req.Pagination),
	})
	if err != nil {
		slog.Error("ListDLQMessages failed", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to list DLQ messages: %v", err)
	}
	messages := make([]*cdcpb.DLQMessage, 0, len(result.Data))
	for _, message := range result.Data {
		messages = append(messages, &cdcpb.DLQMessage{
			Sequence:        message.Message.Sequence,
			Timestamp:       message.Message.Timestamp,
			Subject:         message.Message.Subject,
			Data:            message.Message.Data,
			Headers:         message.Message.Headers,
			Reason:          message.Reason,
			OriginalSubject: message.OriginalSubject,
		})
	}
	return &cdcpb.ListDLQMessagesResponse{Data: messages, Pagination: paginationToProto(result.Pagination)}, nil
}

func protoMessageStatus(status cdcpb.MessageStatus) models.MessageStatus {
	switch status {
	case cdcpb.MessageStatus_MESSAGE_STATUS_SENT:
		return models.MessageStatusSent
	case cdcpb.MessageStatus_MESSAGE_STATUS_UNSENT:
		return models.MessageStatusUnsent
	default:
		return models.MessageStatusAll
	}
}

func paginationPage(p *cdcpb.OffsetPaginationRequest) int {
	if p == nil || p.Page == 0 {
		return 1
	}
	return int(p.Page)
}

func paginationLimit(p *cdcpb.OffsetPaginationRequest) int {
	if p == nil || p.Limit == 0 {
		return 25
	}
	return int(p.Limit)
}

func paginationToProto(p response.PaginationResponse) *cdcpb.OffsetPaginationResponse {
	return &cdcpb.OffsetPaginationResponse{
		TotalRows: p.TotalRows,
		Page:      uint32(p.Page),
		Limit:     uint32(p.Limit),
		HasNext:   p.HasNext,
		HasPrev:   p.HasPrev,
	}
}

func messagesToProto(messages []*interfaces.NATSMessageItem) []*cdcpb.MessageItem {
	result := make([]*cdcpb.MessageItem, 0, len(messages))
	for _, message := range messages {
		result = append(result, &cdcpb.MessageItem{
			Sequence:  message.Sequence,
			Timestamp: strconv.FormatInt(message.Timestamp, 10),
			Subject:   message.Subject,
			Data:      message.Data,
			Headers:   message.Headers,
		})
	}
	return result
}
