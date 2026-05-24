package response

import "github.com/foden/cdc/internal/core/ports"

type CreateFlowResponse struct {
	FlowID string
	Status ports.FlowStatus
}

type GetFlowResponse struct {
	Flow *ports.FlowConfig
}

type ListFlowsResponse struct {
	Flows []*ports.FlowConfig
}

type UpdateFlowResponse struct {
	Flow *ports.FlowConfig
}

type DeleteFlowResponse struct {
	Success bool
}

type PauseFlowResponse struct {
	Status ports.FlowStatus
}

type ResumeFlowResponse struct {
	Status ports.FlowStatus
}

type GetFlowStatsResponse struct {
	Stats *ports.FlowStats
}

type GetFlowTableProgressResponse struct{}
