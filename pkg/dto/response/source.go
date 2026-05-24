package response

import "github.com/foden/cdc/pkg/interfaces"

type CreateSourceResponse struct {
	InstanceID string
}

type GetSourceResponse struct {
	Source *interfaces.SourceConfig
}

type ListSourcesResponse struct {
	Sources []*interfaces.SourceConfig
}

type DeleteSourceResponse struct {
	Success bool
}

type TestSourceConnectionResponse struct {
	Success   bool
	Message   string
	LatencyMs int64
}
