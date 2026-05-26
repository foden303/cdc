package service

import (
	"context"
	"strings"

	"github.com/foden/cdc/internal/core/dto/request"
	"github.com/foden/cdc/internal/core/dto/response"
	"github.com/foden/cdc/internal/core/ports"
	cdcerrors "github.com/foden/cdc/pkg/errors"
)

type SourceService struct {
	store     ports.Store
	discovery ports.Discovery
}

func NewSourceService(store ports.Store, discovery ports.Discovery) *SourceService {
	return &SourceService{store: store, discovery: discovery}
}

func (s *SourceService) Create(ctx context.Context, req request.CreateSourceRequest) (response.CreateSourceResponse, error) {
	if req.Source == nil {
		return response.CreateSourceResponse{}, cdcerrors.ErrSourceConfigRequired
	}
	if req.Source.InstanceID == "" {
		req.Source.InstanceID = defaultInstanceID()
	}
	req.Source.InstanceID = strings.TrimSpace(req.Source.InstanceID)
	req.Source.Name = strings.TrimSpace(req.Source.Name)
	req.Source.Type = strings.TrimSpace(req.Source.Type)
	req.Source.Host = strings.TrimSpace(req.Source.Host)
	req.Source.Username = strings.TrimSpace(req.Source.Username)
	req.Source.Database = strings.TrimSpace(req.Source.Database)
	if err := validateUniqueSource(ctx, s.store, req.Source); err != nil {
		return response.CreateSourceResponse{}, err
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
