package request

import "github.com/foden/cdc/pkg/interfaces"

type CreateSinkRequest struct {
	Sink *interfaces.SinkConfig
}

type GetSinkRequest struct {
	InstanceID string
}

type ListSinksRequest struct{}

type DeleteSinkRequest struct {
	InstanceID string
}

type TestSinkConnectionRequest struct {
	Sink *interfaces.SinkConfig
}
