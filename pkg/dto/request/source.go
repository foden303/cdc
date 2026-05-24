package request

import "github.com/foden/cdc/pkg/interfaces"

type CreateSourceRequest struct {
	Source *interfaces.SourceConfig
}

type GetSourceRequest struct {
	InstanceID string
}

type ListSourcesRequest struct{}

type DeleteSourceRequest struct {
	InstanceID string
}

type TestSourceConnectionRequest struct {
	Source *interfaces.SourceConfig
}
