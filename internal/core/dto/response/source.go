package response

import "github.com/foden/cdc/internal/core/ports"

type CreateSourceResponse struct {
	InstanceID string
}

type GetSourceResponse struct {
	Source *ports.SourceConfig
}

type ListSourcesResponse struct {
	Sources []*ports.SourceConfig
}

type DeleteSourceResponse struct {
	Success bool
}

type TestSourceConnectionResponse struct {
	Success   bool
	Message   string
	LatencyMs int64
}
