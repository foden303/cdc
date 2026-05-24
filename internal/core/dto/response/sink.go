package response

import "github.com/foden/cdc/internal/core/ports"

type CreateSinkResponse struct {
	InstanceID string
}

type GetSinkResponse struct {
	Sink *ports.SinkConfig
}

type ListSinksResponse struct {
	Sinks []*ports.SinkConfig
}

type DeleteSinkResponse struct {
	Success bool
}

type TestSinkConnectionResponse struct {
	Success   bool
	Message   string
	LatencyMs int64
}
