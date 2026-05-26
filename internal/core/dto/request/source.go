package request

import "github.com/foden/cdc/internal/core/ports"

type CreateSourceRequest struct {
	Source *ports.SourceConfig
}

type GetSourceRequest struct {
	InstanceID string
}

type ListSourcesRequest struct{}

type DeleteSourceRequest struct {
	InstanceID string
}

type TestSourceConnectionRequest struct {
	Source *ports.SourceConfig
}
