package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/config"
	"github.com/foden/cdc/pkg/registry"
)

// sourceFingerprint returns a canonical key that uniquely identifies a source's
// connection endpoint: type://host:port/database. Two sources with the same
// fingerprint would read the same replication stream and corrupt data.
func sourceFingerprint(c *config.SourceConfig) string {
	return fmt.Sprintf("%s://%s:%d/%s",
		strings.ToLower(c.Type),
		strings.ToLower(c.Host),
		c.Port,
		strings.ToLower(c.Database),
	)
}

// checkDuplicateSource returns an error if any existing source (excluding
// excludeID) shares the same connection fingerprint as candidate.
func (s *GRPCService) checkDuplicateSource(ctx context.Context, candidate *config.SourceConfig, excludeID string) error {
	existing, err := s.natsClient.GetSourcesConfig(ctx)
	if err != nil {
		slog.Warn("failed to load existing sources for duplicate check, allowing operation", "err", err)
		return nil
	}

	fp := sourceFingerprint(candidate)
	for _, src := range existing {
		if src.InstanceID == excludeID {
			continue
		}
		if sourceFingerprint(src) == fp {
			return fmt.Errorf(
				"duplicate source: instance %q already connects to %s — adding another would corrupt data",
				src.InstanceID, fp,
			)
		}
	}
	return nil
}

func (s *GRPCService) AddSource(ctx context.Context, req *cdcpb.AddSourceRequest) (*cdcpb.AddSourceResponse, error) {
	sCfg, err := toSourceConfig(req.Source)
	if err != nil {
		return nil, statusErrorForAction(err, "add source")
	}

	if err := s.checkDuplicateSource(ctx, sCfg, ""); err != nil {
		return nil, statusErrorForAction(err, "add source")
	}

	src, err := registry.CreateSource(sCfg)
	if err != nil {
		return nil, statusErrorForAction(err, "add source")
	}

	if err := s.engine.AddSource(ctx, src); err != nil {
		return nil, statusErrorForAction(err, "add source")
	}

	rev, err := s.natsClient.SaveSourceConfig(ctx, sCfg)
	if err != nil {
		slog.Error("failed to persist source config", "err", err)
		_ = s.engine.RemoveSource(sCfg.InstanceID)
		return nil, persistedConfigError("source", err)
	}
	s.engine.SetRevision(s.natsClient.SourceConfigKey(sCfg.InstanceID), rev)

	return &cdcpb.AddSourceResponse{InstanceId: sCfg.InstanceID}, nil
}

func (s *GRPCService) RemoveSource(ctx context.Context, req *cdcpb.RemoveSourceRequest) (*cdcpb.RemoveSourceResponse, error) {
	if err := requireSourceInstanceID(req.InstanceId); err != nil {
		return nil, statusErrorForAction(err, "remove source")
	}
	if err := s.engine.RemoveSource(req.InstanceId); err != nil {
		return nil, statusErrorForAction(err, "remove source")
	}
	if err := s.natsClient.RemoveSourceConfig(ctx, req.InstanceId); err != nil {
		slog.Error("failed to remove source config from storage", "err", err)
		return nil, persistedConfigError("source", err)
	}

	return &cdcpb.RemoveSourceResponse{Success: true}, nil
}

func (s *GRPCService) UpdateSource(ctx context.Context, req *cdcpb.UpdateSourceRequest) (*cdcpb.UpdateSourceResponse, error) {
	if err := requireSourceInstanceID(req.Source.InstanceId); err != nil {
		return nil, statusErrorForAction(err, "update source")
	}

	sCfg, err := toSourceConfig(req.Source)
	if err != nil {
		return nil, statusErrorForAction(err, "update source")
	}

	if err := s.checkDuplicateSource(ctx, sCfg, sCfg.InstanceID); err != nil {
		return nil, statusErrorForAction(err, "update source")
	}

	if err := s.engine.RemoveSource(sCfg.InstanceID); err != nil {
		return nil, statusErrorForAction(err, "update source")
	}

	src, err := registry.CreateSource(sCfg)
	if err != nil {
		return nil, statusErrorForAction(err, "update source")
	}
	if err := s.engine.AddSource(ctx, src); err != nil {
		return nil, statusErrorForAction(err, "update source")
	}

	rev, err := s.natsClient.SaveSourceConfig(ctx, sCfg)
	if err != nil {
		slog.Error("failed to persist updated source config", "err", err)
		_ = s.engine.RemoveSource(sCfg.InstanceID)
		return nil, persistedConfigError("source", err)
	}
	s.engine.SetRevision(s.natsClient.SourceConfigKey(sCfg.InstanceID), rev)

	return &cdcpb.UpdateSourceResponse{Success: true}, nil
}
