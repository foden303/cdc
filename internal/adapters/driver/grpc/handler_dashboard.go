package drivergrpc

import (
	"context"
	"log/slog"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/internal/core/dto/request"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CDCService) GetDashboardSummary(
	ctx context.Context,
	_ *cdcpb.GetDashboardSummaryRequest,
) (*cdcpb.GetDashboardSummaryResponse, error) {
	summary, err := s.dashboardService.Summary(ctx, request.DashboardSummaryRequest{})
	if err != nil {
		slog.Error("GetDashboardSummary failed", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to load dashboard summary: %v", err)
	}

	return &cdcpb.GetDashboardSummaryResponse{
		Inventory: &cdcpb.DashboardSystemInventory{
			SourcesCount: int32(summary.Inventory.SourcesCount),
			SinksCount:   int32(summary.Inventory.SinksCount),
			FlowsCount:   int32(summary.Inventory.FlowsCount),
		},
		Telemetry: &cdcpb.DashboardLiveTelemetry{
			Throughput:         summary.Telemetry.Throughput,
			LatencyP99:         summary.Telemetry.LatencyP99,
			ActiveWorkers:      summary.Telemetry.ActiveWorkers,
			ChannelUtilization: summary.Telemetry.ChannelUtilization,
			NatsHealthy:        summary.Telemetry.NATSHealthy,
			ErrorRate:          summary.Telemetry.ErrorRate,
			TotalSyncedEvents:  summary.Telemetry.TotalSyncedEvents,
			FailureCount:       summary.Telemetry.FailureCount,
		},
	}, nil
}
