package service

import (
	"context"
	"time"

	"github.com/foden/cdc/internal/core/dto/request"
	"github.com/foden/cdc/internal/core/dto/response"
	"github.com/foden/cdc/internal/core/ports"
	coreruntime "github.com/foden/cdc/internal/core/runtime"
	"github.com/foden/cdc/version"
)

type MetricsService struct {
	store       ports.Store
	runtimeView *coreruntime.View
	startTime   time.Time
}

func NewMetricsService(store ports.Store, _ ports.FlowManager, runtimeView *coreruntime.View) *MetricsService {
	if runtimeView == nil {
		runtimeView = coreruntime.DefaultView()
	}
	return &MetricsService{
		store:       store,
		runtimeView: runtimeView,
		startTime:   time.Now(),
	}
}

func (s *MetricsService) HealthCheck(_ context.Context, _ request.HealthCheckRequest) response.HealthCheckResponse {
	return response.HealthCheckResponse{
		Status:  "healthy",
		Version: version.Version,
		Uptime:  int64(time.Since(s.startTime).Seconds()),
	}
}

func (s *MetricsService) Stats(ctx context.Context, _ request.GetStatsRequest) (response.GetStatsResponse, error) {
	sources, err := s.store.ListSources(ctx)
	if err != nil {
		return response.GetStatsResponse{}, err
	}
	sinks, err := s.store.ListSinks(ctx)
	if err != nil {
		return response.GetStatsResponse{}, err
	}

	sourceStats := make(map[string]*response.ComponentStats, len(sources))
	sinkStats := make(map[string]*response.ComponentStats, len(sinks))
	for _, src := range sources {
		sourceStats[src.InstanceID] = componentStatsFromRuntime(s.runtimeView.SourceStats(src.InstanceID))
	}
	for _, sk := range sinks {
		sinkStats[sk.InstanceID] = componentStatsFromRuntime(s.runtimeView.SinkStats(sk.InstanceID))
	}

	return response.GetStatsResponse{
		SourceStats: sourceStats,
		SinkStats:   sinkStats,
	}, nil
}

func componentStatsFromRuntime(stats coreruntime.ComponentStatsSnapshot) *response.ComponentStats {
	return &response.ComponentStats{
		SuccessCount: stats.SuccessCount,
		FailureCount: stats.FailureCount,
		LastError:    stats.LastError,
		LastEventAt:  stats.LastEventAt,
		ActiveFlows:  stats.ActiveFlows,
		Throughput:   stats.Throughput,
		ErrorRate:    stats.ErrorRate,
		AvgLatencyMs: stats.AvgLatencyMs,
	}
}
