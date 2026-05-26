package response

type DashboardSummaryResponse struct {
	Inventory DashboardSystemInventoryResponse `json:"inventory"`
	Telemetry DashboardLiveTelemetryResponse   `json:"telemetry"`
}

type DashboardSystemInventoryResponse struct {
	SourcesCount int `json:"sources_count"`
	SinksCount   int `json:"sinks_count"`
	FlowsCount   int `json:"flows_count"`
}

type DashboardLiveTelemetryResponse struct {
	Throughput         float64 `json:"throughput"`
	LatencyP99         float64 `json:"latency_p99"`
	ActiveWorkers      uint32  `json:"active_workers"`
	ChannelUtilization float64 `json:"channel_utilization"`
	NATSHealthy        bool    `json:"nats_healthy"`
	ErrorRate          float64 `json:"error_rate"`
	TotalSyncedEvents  uint64  `json:"total_synced_events"`
	FailureCount       uint64  `json:"failure_count"`
}
