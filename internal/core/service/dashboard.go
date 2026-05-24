package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/foden/cdc/internal/core/dto/request"
	"github.com/foden/cdc/internal/core/dto/response"
	"github.com/foden/cdc/internal/core/ports"
	"github.com/foden/cdc/version"
)

const (
	throughputHistoryLimit = 60
)

type DashboardService struct {
	store       ports.Store
	flowManager ports.FlowManager
	natsClient  ports.NATSClient
	startTime   time.Time

	throughputMu      sync.Mutex
	throughputHistory []response.DashboardThroughputPoint
}

func NewDashboardService(
	store ports.Store,
	flowManager ports.FlowManager,
	natsClient ports.NATSClient,
) *DashboardService {
	return &DashboardService{
		store:       store,
		flowManager: flowManager,
		natsClient:  natsClient,
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
		channelUtil     float64
		runningFlows    int
	)

	for _, flow := range flows {
		if flow.Status != ports.FlowStatusRunning {
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
		activeWorkers += flowStats.RunningWorkers
		if flowStats.WorkerUtilization > channelUtil {
			channelUtil = flowStats.WorkerUtilization
		}
	}

	var latencyP99 float64
	if runningFlows > 0 {
		latencyP99 = totalLatency / float64(runningFlows)
	}

	natsHealthy := false
	if s.natsClient != nil {
		healthCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
		defer cancel()
		if err := s.natsClient.Health(healthCtx); err != nil {
			slog.Warn("dashboard telemetry: nats health unavailable", "err", err)
		} else {
			natsHealthy = true
		}
	}

	return response.DashboardLiveTelemetryResponse{
		Throughput:         totalThroughput,
		LatencyP99:         latencyP99,
		ActiveWorkers:      activeWorkers,
		ChannelUtilization: channelUtil,
		NATSHealthy:        natsHealthy,
		ErrorRate:          0,
		TotalSyncedEvents:  totalEvents,
		FailureCount:       0,
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
