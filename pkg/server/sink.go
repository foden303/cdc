package server

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/config"
	"github.com/foden/cdc/pkg/registry"
)

// sinkFingerprint returns a canonical key that uniquely identifies a sink's
// connection endpoint: type://url1,url2. Two sinks with the same fingerprint
// would write duplicate data to the same target.
func sinkFingerprint(c *config.SinkConfig) string {
	urls := make([]string, len(c.URL))
	copy(urls, c.URL)
	sort.Strings(urls)
	return fmt.Sprintf("%s://%s",
		strings.ToLower(c.Type),
		strings.Join(urls, ","),
	)
}

// checkDuplicateSink returns an error if any existing sink (excluding
// excludeID) shares the same connection fingerprint as candidate.
func (s *GRPCService) checkDuplicateSink(ctx context.Context, candidate *config.SinkConfig, excludeID string) error {
	existing, err := s.natsClient.GetSinksConfig(ctx)
	if err != nil {
		slog.Warn("failed to load existing sinks for duplicate check, allowing operation", "err", err)
		return nil
	}

	fp := sinkFingerprint(candidate)
	for _, snk := range existing {
		if snk.InstanceID == excludeID {
			continue
		}
		if sinkFingerprint(snk) == fp {
			return fmt.Errorf(
				"duplicate sink: instance %q already connects to %s — adding another would cause duplicate writes",
				snk.InstanceID, fp,
			)
		}
	}
	return nil
}

func (s *GRPCService) AddSink(ctx context.Context, req *cdcpb.AddSinkRequest) (*cdcpb.AddSinkResponse, error) {
	sCfg, err := toSinkConfig(req.Sink)
	if err != nil {
		return nil, statusErrorForAction(err, "add sink")
	}

	if err := s.checkDuplicateSink(ctx, sCfg, ""); err != nil {
		return nil, statusErrorForAction(err, "add sink")
	}

	sink, err := registry.CreateSink(sCfg)
	if err != nil {
		return nil, statusErrorForAction(err, "add sink")
	}

	if err := s.engine.AddSink(ctx, sink); err != nil {
		return nil, statusErrorForAction(err, "add sink")
	}

	rev, err := s.natsClient.SaveSinkConfig(ctx, sCfg)
	if err != nil {
		slog.Error("failed to persist sink config", "err", err)
		_ = s.engine.RemoveSink(sCfg.InstanceID)
		return nil, persistedConfigError("sink", err)
	}
	s.engine.SetRevision(s.natsClient.SinkConfigKey(sCfg.InstanceID), rev)

	return &cdcpb.AddSinkResponse{InstanceId: sCfg.InstanceID}, nil
}

func (s *GRPCService) RemoveSink(ctx context.Context, req *cdcpb.RemoveSinkRequest) (*cdcpb.RemoveSinkResponse, error) {
	if err := requireSinkInstanceID(req.InstanceId); err != nil {
		return nil, statusErrorForAction(err, "remove sink")
	}
	if err := s.engine.RemoveSink(req.InstanceId); err != nil {
		return nil, statusErrorForAction(err, "remove sink")
	}
	if err := s.natsClient.RemoveSinkConfig(ctx, req.InstanceId); err != nil {
		slog.Error("failed to remove sink config from storage", "err", err)
		return nil, persistedConfigError("sink", err)
	}

	return &cdcpb.RemoveSinkResponse{Success: true}, nil
}

func (s *GRPCService) UpdateSink(ctx context.Context, req *cdcpb.UpdateSinkRequest) (*cdcpb.UpdateSinkResponse, error) {
	if err := requireSinkInstanceID(req.Sink.InstanceId); err != nil {
		return nil, statusErrorForAction(err, "update sink")
	}

	sCfg, err := toSinkConfig(req.Sink)
	if err != nil {
		return nil, statusErrorForAction(err, "update sink")
	}

	if err := s.checkDuplicateSink(ctx, sCfg, sCfg.InstanceID); err != nil {
		return nil, statusErrorForAction(err, "update sink")
	}

	if err := s.engine.RemoveSink(sCfg.InstanceID); err != nil {
		return nil, statusErrorForAction(err, "update sink")
	}

	sink, err := registry.CreateSink(sCfg)
	if err != nil {
		return nil, statusErrorForAction(err, "update sink")
	}
	if err := s.engine.AddSink(ctx, sink); err != nil {
		return nil, statusErrorForAction(err, "update sink")
	}

	rev, err := s.natsClient.SaveSinkConfig(ctx, sCfg)
	if err != nil {
		slog.Error("failed to persist updated sink config", "err", err)
		_ = s.engine.RemoveSink(sCfg.InstanceID)
		return nil, persistedConfigError("sink", err)
	}
	s.engine.SetRevision(s.natsClient.SinkConfigKey(sCfg.InstanceID), rev)

	return &cdcpb.UpdateSinkResponse{Success: true}, nil
}
