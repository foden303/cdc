package request

import "github.com/foden/cdc/internal/core/ports"

type CreateFlowRequest struct {
	Flow *ports.FlowConfig
}

type GetFlowRequest struct {
	FlowID string
}

type ListFlowsRequest struct{}

type UpdateFlowRequest struct {
	Flow *ports.FlowConfig
}

type DeleteFlowRequest struct {
	FlowID string
}

type PauseFlowRequest struct {
	FlowID string
}

type ResumeFlowRequest struct {
	FlowID string
}

type GetFlowStatsRequest struct {
	FlowID string
}

type GetFlowTableProgressRequest struct {
	FlowID string
}
