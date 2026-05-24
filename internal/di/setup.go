package di

import (
	drivergrpc "github.com/foden/cdc/internal/adapters/driver/grpc"
	"github.com/foden/cdc/internal/core/ports"
	"github.com/foden/cdc/internal/core/service"
)

type Resources struct {
	Store       ports.Store
	FlowManager ports.FlowManager
	Registry    ports.Registry
	Discovery   ports.Discovery
	NATSClient  ports.NATSClient
}

func SetupDependencies(resources Resources) *Container {
	container := &Container{
		Store:            resources.Store,
		FlowManager:      resources.FlowManager,
		Registry:         resources.Registry,
		Discovery:        resources.Discovery,
		NATSClient:       resources.NATSClient,
		CDCService:       drivergrpc.NewCDCService(resources.Store, resources.FlowManager, resources.Registry, resources.Discovery, resources.NATSClient),
		DashboardService: service.NewDashboardService(resources.Store, resources.FlowManager),
	}
	GlobalContainer = container
	return container
}
