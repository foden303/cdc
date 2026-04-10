package server

import (
	"context"
	"fmt"
	"log/slog"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/config"
	"github.com/foden/cdc/pkg/registry"
)

func (s *GRPCService) AddSource(ctx context.Context, req *cdcpb.AddSourceRequest) (*cdcpb.AddSourceResponse, error) {
	sCfg, err := toSourceConfig(req.Source)
	if err != nil {
		return nil, fmt.Errorf("invalid source config: %w", err)
	}

	src, err := registry.CreateSource(sCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	if err := s.engine.AddSource(ctx, src); err != nil {
		return nil, err
	}

	// Persist updated configuration
	sCfg.InstanceID = src.InstanceID()
	updated := false
	for i := range s.appCfg.Sources {
		if s.appCfg.Sources[i].InstanceID == sCfg.InstanceID {
			s.appCfg.Sources[i] = sCfg
			updated = true
			break
		}
	}
	if !updated {
		s.appCfg.Sources = append(s.appCfg.Sources, sCfg)
	}

	if err := s.natsClient.SaveConfig(ctx, s.appCfg); err != nil {
		slog.Error("failed to persist config", "err", err)
	}

	return &cdcpb.AddSourceResponse{InstanceId: src.InstanceID()}, nil
}

func (s *GRPCService) RemoveSource(ctx context.Context, req *cdcpb.RemoveSourceRequest) (*cdcpb.RemoveSourceResponse, error) {
	if err := s.engine.RemoveSource(req.InstanceId); err != nil {
		return nil, err
	}

	// Persist updated configuration
	var newSources []*config.SourceConfig
	for i := range s.appCfg.Sources {
		if s.appCfg.Sources[i].InstanceID != req.InstanceId {
			newSources = append(newSources, s.appCfg.Sources[i])
		}
	}
	s.appCfg.Sources = newSources
	if err := s.natsClient.SaveConfig(ctx, s.appCfg); err != nil {
		slog.Error("failed to persist config", "err", err)
	}

	return &cdcpb.RemoveSourceResponse{Success: true}, nil
}
func (s *GRPCService) UpdateSource(ctx context.Context, req *cdcpb.UpdateSourceRequest) (*cdcpb.UpdateSourceResponse, error) {
	if req.Source.InstanceId == "" {
		return nil, fmt.Errorf("instance_id is required for update")
	}

	// 1. Convert to internal config
	sCfg, err := toSourceConfig(req.Source)
	if err != nil {
		return nil, fmt.Errorf("invalid source config: %w", err)
	}

	// 2. Remove old instance from engine
	if err := s.engine.RemoveSource(req.Source.InstanceId); err != nil {
		return nil, fmt.Errorf("failed to stop old source: %w", err)
	}

	// 3. Create and add new instance
	src, err := registry.CreateSource(sCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new source version: %w", err)
	}

	if err := s.engine.AddSource(ctx, src); err != nil {
		return nil, fmt.Errorf("failed to start new source version: %w", err)
	}

	// 4. Update and persist config
	for i := range s.appCfg.Sources {
		if s.appCfg.Sources[i].InstanceID == req.Source.InstanceId {
			s.appCfg.Sources[i] = sCfg
			break
		}
	}

	if err := s.natsClient.SaveConfig(ctx, s.appCfg); err != nil {
		slog.Error("failed to persist updated source config", "err", err)
	}

	return &cdcpb.UpdateSourceResponse{Success: true}, nil
}
