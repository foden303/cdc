package service

import (
	"context"

	"github.com/foden/cdc/internal/core/dto/request"
	"github.com/foden/cdc/internal/core/dto/response"
	"github.com/foden/cdc/internal/core/ports"
	cdcerrors "github.com/foden/cdc/pkg/errors"
)

type SinkService struct {
	store     ports.Store
	discovery ports.Discovery
}

func NewSinkService(store ports.Store, discovery ports.Discovery) *SinkService {
	return &SinkService{store: store, discovery: discovery}
}

func (s *SinkService) Create(ctx context.Context, req request.CreateSinkRequest) (response.CreateSinkResponse, error) {
	if req.Sink == nil {
		return response.CreateSinkResponse{}, cdcerrors.ErrSinkConfigRequired
	}
	if req.Sink.InstanceID == "" {
		req.Sink.InstanceID = defaultInstanceID()
	}
	if err := s.store.PutSink(ctx, req.Sink); err != nil {
		return response.CreateSinkResponse{}, err
	}
	return response.CreateSinkResponse{InstanceID: req.Sink.InstanceID}, nil
}

func (s *SinkService) Get(ctx context.Context, req request.GetSinkRequest) (response.GetSinkResponse, error) {
	sink, err := s.store.GetSink(ctx, req.InstanceID)
	if err != nil {
		return response.GetSinkResponse{}, err
	}
	return response.GetSinkResponse{Sink: sink}, nil
}

func (s *SinkService) List(ctx context.Context, _ request.ListSinksRequest) (response.ListSinksResponse, error) {
	sinks, err := s.store.ListSinks(ctx)
	if err != nil {
		return response.ListSinksResponse{}, err
	}
	return response.ListSinksResponse{Sinks: sinks}, nil
}

func (s *SinkService) Delete(ctx context.Context, req request.DeleteSinkRequest) (response.DeleteSinkResponse, error) {
	if err := s.store.DeleteSink(ctx, req.InstanceID); err != nil {
		return response.DeleteSinkResponse{}, err
	}
	return response.DeleteSinkResponse{Success: true}, nil
}

func (s *SinkService) TestConnection(
	ctx context.Context,
	req request.TestSinkConnectionRequest,
) (response.TestSinkConnectionResponse, error) {
	if req.Sink == nil {
		return response.TestSinkConnectionResponse{}, cdcerrors.ErrSinkConfigRequired
	}
	latencyMs, err := s.discovery.TestSinkConnection(ctx, req.Sink)
	if err != nil {
		return response.TestSinkConnectionResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return response.TestSinkConnectionResponse{
		Success:   true,
		Message:   "connection successful",
		LatencyMs: latencyMs,
	}, nil
}
