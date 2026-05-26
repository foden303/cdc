package response

type HealthCheckResponse struct {
	Status  string
	Version string
	Uptime  int64
}

type ComponentStats struct {
	SuccessCount uint64
	FailureCount uint64
	LastError    string
	PartitionLag map[int32]uint64
	LastEventAt  int64
	ActiveFlows  int32
	Throughput   float64
	ErrorRate    float64
	AvgLatencyMs int64
}

type GetStatsResponse struct {
	SourceStats map[string]*ComponentStats
	SinkStats   map[string]*ComponentStats
}
