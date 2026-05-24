package response

import "github.com/foden/cdc/pkg/interfaces"

type CreateSinkResponse struct {
	InstanceID string
}

type GetSinkResponse struct {
	Sink *interfaces.SinkConfig
}

type ListSinksResponse struct {
	Sinks []*interfaces.SinkConfig
}

type DeleteSinkResponse struct {
	Success bool
}

type TestSinkConnectionResponse struct {
	Success   bool
	Message   string
	LatencyMs int64
}
