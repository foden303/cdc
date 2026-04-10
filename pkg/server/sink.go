package server

import (
	"context"
	"fmt"
	"log/slog"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/config"
	"github.com/foden/cdc/pkg/registry"
)

func (s *GRPCService) AddSink(ctx context.Context, req *cdcpb.AddSinkRequest) (*cdcpb.AddSinkResponse, error) {
	sCfg, err := toSinkConfig(req.Sink)
	if err != nil {
		return nil, fmt.Errorf("invalid sink config: %w", err)
	}

	sink, err := registry.CreateSink(sCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create sink: %w", err)
	}

	s.engine.AddSink(sink)

	// Persist updated configuration
	updated := false
	for i := range s.appCfg.Sinks {
		if s.appCfg.Sinks[i].InstanceID == sCfg.InstanceID {
			s.appCfg.Sinks[i] = sCfg
			updated = true
			break
		}
	}
	if !updated {
		s.appCfg.Sinks = append(s.appCfg.Sinks, sCfg)
	}

	if err := s.natsClient.SaveConfig(ctx, s.appCfg); err != nil {
		slog.Error("failed to persist config", "err", err)
	}

	return &cdcpb.AddSinkResponse{InstanceId: sink.InstanceID()}, nil
}

func (s *GRPCService) RemoveSink(ctx context.Context, req *cdcpb.RemoveSinkRequest) (*cdcpb.RemoveSinkResponse, error) {
	if err := s.engine.RemoveSink(req.InstanceId); err != nil {
		return nil, err
	}

	// Persist updated configuration
	var newSinks []*config.SinkConfig
	for i := range s.appCfg.Sinks {
		if s.appCfg.Sinks[i].InstanceID != req.InstanceId {
			newSinks = append(newSinks, s.appCfg.Sinks[i])
		}
	}
	s.appCfg.Sinks = newSinks
	if err := s.natsClient.SaveConfig(ctx, s.appCfg); err != nil {
		slog.Error("failed to persist config", "err", err)
	}

	return &cdcpb.RemoveSinkResponse{Success: true}, nil
}
func (s *GRPCService) UpdateSink(ctx context.Context, req *cdcpb.UpdateSinkRequest) (*cdcpb.UpdateSinkResponse, error) {
	if req.Sink.InstanceId == "" {
		return nil, fmt.Errorf("instance_id is required for update")
	}

	// 1. Convert to internal config
	sCfg, err := toSinkConfig(req.Sink)
	if err != nil {
		return nil, fmt.Errorf("invalid sink config: %w", err)
	}

	// 2. Remove old instance from engine
	if err := s.engine.RemoveSink(req.Sink.InstanceId); err != nil {
		return nil, fmt.Errorf("failed to stop old sink: %w", err)
	}

	// 3. Create and add new instance
	sink, err := registry.CreateSink(sCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new sink version: %w", err)
	}

	s.engine.AddSink(sink)

	// 4. Update and persist config
	for i := range s.appCfg.Sinks {
		if s.appCfg.Sinks[i].InstanceID == req.Sink.InstanceId {
			s.appCfg.Sinks[i] = sCfg
			break
		}
	}

	if err := s.natsClient.SaveConfig(ctx, s.appCfg); err != nil {
		slog.Error("failed to persist updated sink config", "err", err)
	}

	return &cdcpb.UpdateSinkResponse{Success: true}, nil
}
