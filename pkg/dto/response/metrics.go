package response

type HealthCheckResponse struct {
	Status string
	Uptime int64
}

type ComponentStats struct {
	SuccessCount uint64
	FailureCount uint64
	LastError    string
	PartitionLag map[int32]uint64
}

type GetStatsResponse struct {
	SourceStats map[string]*ComponentStats
	SinkStats   map[string]*ComponentStats
}

type SourcePerformance struct {
	SourceID   string
	Throughput float64
	ErrorRate  float64
}

type SinkPerformance struct {
	SinkID     string
	Throughput float64
	AvgLatency float64
}

type GetPerformanceMetricsResponse struct {
	Throughput    float64
	LatencyP99    float64
	ActiveWorkers uint32
	ErrorRate     float64
	Sources       map[string]*SourcePerformance
	Sinks         map[string]*SinkPerformance
}
