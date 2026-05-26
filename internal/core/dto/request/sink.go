package request

import "github.com/foden/cdc/internal/core/ports"

type CreateSinkRequest struct {
	Sink *ports.SinkConfig
}

type GetSinkRequest struct {
	InstanceID string
}

type ListSinksRequest struct{}

type DeleteSinkRequest struct {
	InstanceID string
}

type TestSinkConnectionRequest struct {
	Sink *ports.SinkConfig
}
