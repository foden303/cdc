package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/foden/cdc/pkg/dto/request"
	"github.com/foden/cdc/pkg/dto/response"
	"github.com/foden/cdc/pkg/interfaces"
	"github.com/foden/cdc/version"
)

const (
	throughputHistoryLimit = 60
	fallbackWorkerCount    = 4
)

type DashboardService struct {
	store       interfaces.Store
	flowManager interfaces.FlowManager
	startTime   time.Time

	throughputMu      sync.Mutex
	throughputHistory []response.DashboardThroughputPoint
}

func NewDashboardService(
	store interfaces.Store,
	flowManager interfaces.FlowManager,
) *DashboardService {
	return &DashboardService{
		store:       store,
		flowManager: flowManager,
		startTime:   time.Now(),
	}
}

func (s *DashboardService) Health(_ context.Context, _ request.DashboardHealthRequest) response.DashboardHealthResponse {
	return response.DashboardHealthResponse{
		Status:  "healthy",
		Version: version.Version,
		Uptime:  int64(time.Since(s.startTime).Seconds()),
	}
}

func (s *DashboardService) SystemInventory(
	ctx context.Context,
	_ request.DashboardSystemInventoryRequest,
) (response.DashboardSystemInventoryResponse, error) {
	sources, err := s.store.ListSources(ctx)
	if err != nil {
		return response.DashboardSystemInventoryResponse{}, err
	}
	sinks, err := s.store.ListSinks(ctx)
	if err != nil {
		return response.DashboardSystemInventoryResponse{}, err
	}
	flows, err := s.store.ListFlows(ctx)
	if err != nil {
		return response.DashboardSystemInventoryResponse{}, err
	}

	return response.DashboardSystemInventoryResponse{
		SourcesCount: len(sources),
		SinksCount:   len(sinks),
		FlowsCount:   len(flows),
	}, nil
}

func (s *DashboardService) LiveTelemetry(
	ctx context.Context,
	_ request.DashboardLiveTelemetryRequest,
) (response.DashboardLiveTelemetryResponse, error) {
	flows, err := s.store.ListFlows(ctx)
	if err != nil {
		return response.DashboardLiveTelemetryResponse{}, err
	}

	var (
		totalThroughput float64
		totalLatency    float64
		totalEvents     uint64
		activeWorkers   uint32
		runningFlows    int
	)

	for _, flow := range flows {
		if flow.Status != interfaces.FlowStatusRunning {
			continue
		}
		runningFlows++

		flowStats, statsErr := s.flowManager.GetFlowStats(ctx, flow.FlowID)
		if statsErr != nil {
			slog.Warn("dashboard telemetry: flow stats unavailable", "flow_id", flow.FlowID, "err", statsErr)
			continue
		}

		totalThroughput += flowStats.EventsPerSecond
		totalLatency += float64(flowStats.ReplicationLagMs)
		totalEvents += flowStats.TotalEventsProcessed
		activeWorkers += dashboardWorkerCount(flow)
	}

	var latencyP99 float64
	if runningFlows > 0 {
		latencyP99 = totalLatency / float64(runningFlows)
	}

	return response.DashboardLiveTelemetryResponse{
		Throughput:        totalThroughput,
		LatencyP99:        latencyP99,
		ActiveWorkers:     activeWorkers,
		ErrorRate:         0,
		TotalSyncedEvents: totalEvents,
		FailureCount:      0,
	}, nil
}

func (s *DashboardService) ThroughputOverTime(
	ctx context.Context,
	_ request.DashboardThroughputOverTimeRequest,
) (response.DashboardThroughputOverTimeResponse, error) {
	telemetry, err := s.LiveTelemetry(ctx, request.DashboardLiveTelemetryRequest{})
	if err != nil {
		return response.DashboardThroughputOverTimeResponse{}, err
	}

	point := response.DashboardThroughputPoint{
		Timestamp:  time.Now().UnixMilli(),
		Throughput: telemetry.Throughput,
	}

	s.throughputMu.Lock()
	s.throughputHistory = append(s.throughputHistory, point)
	if len(s.throughputHistory) > throughputHistoryLimit {
		s.throughputHistory = s.throughputHistory[len(s.throughputHistory)-throughputHistoryLimit:]
	}
	points := append([]response.DashboardThroughputPoint(nil), s.throughputHistory...)
	s.throughputMu.Unlock()

	return response.DashboardThroughputOverTimeResponse{Points: points}, nil
}

func dashboardWorkerCount(flow *interfaces.FlowConfig) uint32 {
	if flow.Options == nil {
		return fallbackWorkerCount
	}
	if flow.Options.PoolSize > 0 {
		return uint32(flow.Options.PoolSize)
	}
	if flow.Options.PartitionCount > 0 {
		return uint32(flow.Options.PartitionCount)
	}
	return fallbackWorkerCount
}
