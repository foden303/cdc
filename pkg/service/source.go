package service

import (
	"context"

	"github.com/foden/cdc/pkg/dto/request"
	"github.com/foden/cdc/pkg/dto/response"
	cdcerrors "github.com/foden/cdc/pkg/errors"
	"github.com/foden/cdc/pkg/interfaces"
)

type SourceService struct {
	store     interfaces.Store
	discovery interfaces.Discovery
}

func NewSourceService(store interfaces.Store, discovery interfaces.Discovery) *SourceService {
	return &SourceService{store: store, discovery: discovery}
}

func (s *SourceService) Create(ctx context.Context, req request.CreateSourceRequest) (response.CreateSourceResponse, error) {
	if req.Source == nil {
		return response.CreateSourceResponse{}, cdcerrors.ErrSourceConfigRequired
	}
	if req.Source.InstanceID == "" {
		req.Source.InstanceID = defaultInstanceID()
	}
	if err := s.store.PutSource(ctx, req.Source); err != nil {
		return response.CreateSourceResponse{}, err
	}
	return response.CreateSourceResponse{InstanceID: req.Source.InstanceID}, nil
}

func (s *SourceService) Get(ctx context.Context, req request.GetSourceRequest) (response.GetSourceResponse, error) {
	source, err := s.store.GetSource(ctx, req.InstanceID)
	if err != nil {
		return response.GetSourceResponse{}, err
	}
	return response.GetSourceResponse{Source: source}, nil
}

func (s *SourceService) List(ctx context.Context, _ request.ListSourcesRequest) (response.ListSourcesResponse, error) {
	sources, err := s.store.ListSources(ctx)
	if err != nil {
		return response.ListSourcesResponse{}, err
	}
	return response.ListSourcesResponse{Sources: sources}, nil
}

func (s *SourceService) Delete(ctx context.Context, req request.DeleteSourceRequest) (response.DeleteSourceResponse, error) {
	if err := s.store.DeleteSource(ctx, req.InstanceID); err != nil {
		return response.DeleteSourceResponse{}, err
	}
	return response.DeleteSourceResponse{Success: true}, nil
}

func (s *SourceService) TestConnection(
	ctx context.Context,
	req request.TestSourceConnectionRequest,
) (response.TestSourceConnectionResponse, error) {
	if req.Source == nil {
		return response.TestSourceConnectionResponse{}, cdcerrors.ErrSourceConfigRequired
	}
	latencyMs, err := s.discovery.TestSourceConnection(ctx, req.Source)
	if err != nil {
		return response.TestSourceConnectionResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return response.TestSourceConnectionResponse{
		Success:   true,
		Message:   "connection successful",
		LatencyMs: latencyMs,
	}, nil
}
