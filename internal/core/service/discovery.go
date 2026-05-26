package service

import (
	"context"

	"github.com/foden/cdc/internal/core/dto/request"
	"github.com/foden/cdc/internal/core/dto/response"
	"github.com/foden/cdc/internal/core/ports"
)

type DiscoveryService struct {
	store     ports.Store
	discovery ports.Discovery
}

func NewDiscoveryService(store ports.Store, discovery ports.Discovery) *DiscoveryService {
	return &DiscoveryService{store: store, discovery: discovery}
}

func (s *DiscoveryService) DiscoverSourceTables(
	ctx context.Context,
	req request.DiscoverTablesRequest,
) (response.DiscoverTablesResponse, error) {
	source, err := s.store.GetSource(ctx, req.SourceID)
	if err != nil {
		return response.DiscoverTablesResponse{}, err
	}
	tables, err := s.discovery.DiscoverSourceTables(ctx, source)
	if err != nil {
		return response.DiscoverTablesResponse{}, err
	}
	return response.DiscoverTablesResponse{Tables: tables}, nil
}

func (s *DiscoveryService) DiscoverSinkTables(
	ctx context.Context,
	req request.DiscoverSinkTablesRequest,
) (response.DiscoverSinkTablesResponse, error) {
	sink, err := s.store.GetSink(ctx, req.SinkID)
	if err != nil {
		return response.DiscoverSinkTablesResponse{}, err
	}
	tables, err := s.discovery.DiscoverSinkTables(ctx, sink)
	if err != nil {
		return response.DiscoverSinkTablesResponse{}, err
	}
	return response.DiscoverSinkTablesResponse{Tables: tables}, nil
}
