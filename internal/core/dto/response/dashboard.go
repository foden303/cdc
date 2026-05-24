package response

type DashboardHealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptime"`
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

type DashboardThroughputPoint struct {
	Timestamp  int64   `json:"timestamp"`
	Throughput float64 `json:"throughput"`
}

type DashboardThroughputOverTimeResponse struct {
	Points []DashboardThroughputPoint `json:"points"`
}
