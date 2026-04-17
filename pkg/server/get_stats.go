package server

import (
	"context"

	cdcpb "github.com/foden/cdc/api/proto/v1"
)

func (s *GRPCService) GetStats(_ context.Context, _ *cdcpb.GetStatsRequest) (*cdcpb.GetStatsResponse, error) {
	srcStats, snkStats := s.engine.GetStats()

	resp := &cdcpb.GetStatsResponse{
		SourceStats: make(map[string]*cdcpb.ComponentStats),
		SinkStats:   make(map[string]*cdcpb.ComponentStats),
	}

	for k, v := range srcStats {
		resp.SourceStats[k] = &cdcpb.ComponentStats{
			SuccessCount: v.SuccessCount,
			FailureCount: v.FailureCount,
			LastError:    v.LastError,
		}
	}

	for k, v := range snkStats {
		resp.SinkStats[k] = &cdcpb.ComponentStats{
			SuccessCount: v.SuccessCount,
			FailureCount: v.FailureCount,
			LastError:    v.LastError,
			PartitionLag: v.PartitionLag,
		}
	}

	return resp, nil
}
