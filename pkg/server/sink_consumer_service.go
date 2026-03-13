package server

import (
	"context"
	"fmt"
	"log/slog"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/config"
	"github.com/foden/cdc/pkg/sink_consumer"
)

// SinkConsumerServiceHandler implements gRPC service for sink consumer management
type SinkConsumerServiceHandler struct {
	cdcpb.UnimplementedSinkConsumerServiceServer
	manager *sink_consumer.Manager
}

// NewSinkConsumerServiceHandler creates a new sink consumer service handler
func NewSinkConsumerServiceHandler(manager *sink_consumer.Manager) *SinkConsumerServiceHandler {
	return &SinkConsumerServiceHandler{
		manager: manager,
	}
}

// CreateConsumer creates and starts a new sink consumer
func (h *SinkConsumerServiceHandler) CreateConsumer(ctx context.Context, req *cdcpb.CreateConsumerRequest) (*cdcpb.CreateConsumerResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("consumer name is required")
	}
	if req.TopicName == "" {
		return nil, fmt.Errorf("topic_name is required")
	}
	if len(req.Sinks) == 0 {
		return nil, fmt.Errorf("at least one sink is required")
	}

	// Convert proto sink configs to internal config
	sinkSpecs := make([]sink_consumer.SinkSpec, len(req.Sinks))
	for i, pbSink := range req.Sinks {
		sinkCfg := &config.SinkConfig{
			Type: pbSink.Type,
			URL:  []string{}, // Will be populated from config map
		}

		// Extract URLs if present
		if urls, ok := pbSink.Config["url"]; ok {
			sinkCfg.URL = append(sinkCfg.URL, urls)
		}

		if indexPrefix, ok := pbSink.Config["index_prefix"]; ok {
			sinkCfg.IndexPrefix = indexPrefix
		}

		sinkSpecs[i] = sink_consumer.SinkSpec{
			Type:   pbSink.Type,
			Config: sinkCfg,
		}
	}

	// Convert partitions
	partitions := make([]int, len(req.Partitions))
	for i, p := range req.Partitions {
		partitions[i] = int(p)
	}

	// Create internal config
	batchSize := int(req.BatchSize)
	if batchSize <= 0 {
		batchSize = 500
	}

	flushIntervalMs := int(req.FlushIntervalMs)
	if flushIntervalMs <= 0 {
		flushIntervalMs = 1000
	}

	internalConfig := sink_consumer.CreateConsumerConfig{
		Name:            req.Name,
		TopicName:       req.TopicName,
		Partitions:      partitions,
		Sinks:           sinkSpecs,
		BatchSize:       batchSize,
		FlushIntervalMs: flushIntervalMs,
	}

	// Create consumer
	consumer, err := h.manager.Create(ctx, internalConfig)
	if err != nil {
		slog.Error("failed to create sink consumer", "err", err)
		return nil, err
	}

	// Get stats
	stats := consumer.GetStats()
	pbStats := convertStatsToProto(stats)

	return &cdcpb.CreateConsumerResponse{
		Id:    consumer.ID,
		Name:  consumer.Name,
		Stats: pbStats,
	}, nil
}

// ListConsumers returns all active sink consumers
func (h *SinkConsumerServiceHandler) ListConsumers(ctx context.Context, _ *cdcpb.ListConsumersRequest) (*cdcpb.ListConsumersResponse, error) {
	consumers := h.manager.List()

	pbConsumers := make([]*cdcpb.ProtoConsumerStats, len(consumers))
	for i, consumer := range consumers {
		stats := consumer.GetStats()
		pbConsumers[i] = convertStatsToProto(stats)
	}

	return &cdcpb.ListConsumersResponse{
		Consumers: pbConsumers,
		Total:     int32(len(pbConsumers)),
	}, nil
}

// GetConsumer returns details of a specific sink consumer
func (h *SinkConsumerServiceHandler) GetConsumer(ctx context.Context, req *cdcpb.GetConsumerRequest) (*cdcpb.ProtoConsumerStats, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("consumer id is required")
	}

	consumer := h.manager.Get(req.Id)
	if consumer == nil {
		return nil, fmt.Errorf("consumer %s not found", req.Id)
	}

	stats := consumer.GetStats()
	return convertStatsToProto(stats), nil
}

// DeleteConsumer stops and deletes a sink consumer
func (h *SinkConsumerServiceHandler) DeleteConsumer(ctx context.Context, req *cdcpb.DeleteConsumerRequest) (*cdcpb.DeleteConsumerResponse, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("consumer id is required")
	}

	if err := h.manager.Delete(req.Id); err != nil {
		slog.Error("failed to delete sink consumer", "consumer_id", req.Id, "err", err)
		return nil, err
	}

	return &cdcpb.DeleteConsumerResponse{
		Message: fmt.Sprintf("consumer %s deleted successfully", req.Id),
	}, nil
}

// PauseConsumer pauses a sink consumer
func (h *SinkConsumerServiceHandler) PauseConsumer(ctx context.Context, req *cdcpb.PauseConsumerRequest) (*cdcpb.ProtoConsumerStats, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("consumer id is required")
	}

	if err := h.manager.Pause(req.Id); err != nil {
		slog.Error("failed to pause sink consumer", "consumer_id", req.Id, "err", err)
		return nil, err
	}

	consumer := h.manager.Get(req.Id)
	if consumer == nil {
		return nil, fmt.Errorf("consumer %s not found", req.Id)
	}

	stats := consumer.GetStats()
	return convertStatsToProto(stats), nil
}

// ResumeConsumer resumes a paused sink consumer
func (h *SinkConsumerServiceHandler) ResumeConsumer(ctx context.Context, req *cdcpb.ResumeConsumerRequest) (*cdcpb.ProtoConsumerStats, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("consumer id is required")
	}

	if err := h.manager.Resume(req.Id); err != nil {
		slog.Error("failed to resume sink consumer", "consumer_id", req.Id, "err", err)
		return nil, err
	}

	consumer := h.manager.Get(req.Id)
	if consumer == nil {
		return nil, fmt.Errorf("consumer %s not found", req.Id)
	}

	stats := consumer.GetStats()
	return convertStatsToProto(stats), nil
}

// convertStatsToProto converts internal ConsumerStats to proto ProtoConsumerStats
func convertStatsToProto(stats *sink_consumer.ConsumerStats) *cdcpb.ProtoConsumerStats {
	lastOffset := make(map[int32]uint64)
	for partID, offset := range stats.LastOffset {
		lastOffset[int32(partID)] = offset
	}

	upSince := int64(0)
	if stats.UpSince != nil {
		upSince = stats.UpSince.Unix()
	}

	return &cdcpb.ProtoConsumerStats{
		Id:             stats.ID,
		Name:           stats.Name,
		TopicName:      stats.TopicName,
		Status:         stats.Status,
		TotalProcessed: stats.TotalProcessed,
		TotalErrors:    stats.TotalErrors,
		LastOffset:     lastOffset,
		LastUpdated:    stats.LastUpdated.Unix(),
		UpSince:        upSince,
		Sinks:          stats.Sinks,
	}
}
