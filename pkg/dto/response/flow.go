package response

import "github.com/foden/cdc/pkg/interfaces"

type CreateFlowResponse struct {
	FlowID string
	Status interfaces.FlowStatus
}

type GetFlowResponse struct {
	Flow *interfaces.FlowConfig
}

type ListFlowsResponse struct {
	Flows []*interfaces.FlowConfig
}

type UpdateFlowResponse struct {
	Flow *interfaces.FlowConfig
}

type DeleteFlowResponse struct {
	Success bool
}

type PauseFlowResponse struct {
	Status interfaces.FlowStatus
}

type ResumeFlowResponse struct {
	Status interfaces.FlowStatus
}

type GetFlowStatsResponse struct {
	Stats *interfaces.FlowStats
}

type GetFlowTableProgressResponse struct{}
