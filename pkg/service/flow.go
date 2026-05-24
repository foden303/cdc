package service

import (
	"context"

	"github.com/foden/cdc/pkg/dto/request"
	"github.com/foden/cdc/pkg/dto/response"
	"github.com/foden/cdc/pkg/interfaces"
)

type FlowService struct {
	flowManager interfaces.FlowManager
}

func NewFlowService(flowManager interfaces.FlowManager) *FlowService {
	return &FlowService{flowManager: flowManager}
}

func (s *FlowService) Create(ctx context.Context, req request.CreateFlowRequest) (response.CreateFlowResponse, error) {
	flow, err := s.flowManager.CreateFlow(ctx, req.Flow)
	if err != nil {
		return response.CreateFlowResponse{}, err
	}
	return response.CreateFlowResponse{FlowID: flow.FlowID, Status: flow.Status}, nil
}

func (s *FlowService) Get(ctx context.Context, req request.GetFlowRequest) (response.GetFlowResponse, error) {
	flow, err := s.flowManager.GetFlow(ctx, req.FlowID)
	if err != nil {
		return response.GetFlowResponse{}, err
	}
	return response.GetFlowResponse{Flow: flow}, nil
}

func (s *FlowService) List(ctx context.Context, _ request.ListFlowsRequest) (response.ListFlowsResponse, error) {
	flows, err := s.flowManager.ListFlows(ctx)
	if err != nil {
		return response.ListFlowsResponse{}, err
	}
	return response.ListFlowsResponse{Flows: flows}, nil
}

func (s *FlowService) Update(ctx context.Context, req request.UpdateFlowRequest) (response.UpdateFlowResponse, error) {
	flow, err := s.flowManager.UpdateFlow(ctx, req.Flow)
	if err != nil {
		return response.UpdateFlowResponse{}, err
	}
	return response.UpdateFlowResponse{Flow: flow}, nil
}

func (s *FlowService) Delete(ctx context.Context, req request.DeleteFlowRequest) (response.DeleteFlowResponse, error) {
	if err := s.flowManager.DeleteFlow(ctx, req.FlowID); err != nil {
		return response.DeleteFlowResponse{}, err
	}
	return response.DeleteFlowResponse{Success: true}, nil
}

func (s *FlowService) Pause(ctx context.Context, req request.PauseFlowRequest) (response.PauseFlowResponse, error) {
	flow, err := s.flowManager.PauseFlow(ctx, req.FlowID)
	if err != nil {
		return response.PauseFlowResponse{}, err
	}
	return response.PauseFlowResponse{Status: flow.Status}, nil
}

func (s *FlowService) Resume(ctx context.Context, req request.ResumeFlowRequest) (response.ResumeFlowResponse, error) {
	flow, err := s.flowManager.ResumeFlow(ctx, req.FlowID)
	if err != nil {
		return response.ResumeFlowResponse{}, err
	}
	return response.ResumeFlowResponse{Status: flow.Status}, nil
}

func (s *FlowService) Stats(ctx context.Context, req request.GetFlowStatsRequest) (response.GetFlowStatsResponse, error) {
	stats, err := s.flowManager.GetFlowStats(ctx, req.FlowID)
	if err != nil {
		return response.GetFlowStatsResponse{}, err
	}
	return response.GetFlowStatsResponse{Stats: stats}, nil
}

func (s *FlowService) TableProgress(
	_ context.Context,
	_ request.GetFlowTableProgressRequest,
) (response.GetFlowTableProgressResponse, error) {
	return response.GetFlowTableProgressResponse{}, nil
}
