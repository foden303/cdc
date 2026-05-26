package di

import (
	drivergrpc "github.com/foden/cdc/internal/adapters/driver/grpc"
	"github.com/foden/cdc/internal/core/ports"
	coreruntime "github.com/foden/cdc/internal/core/runtime"
)

type Resources struct {
	Store       ports.Store
	FlowManager ports.FlowManager
	Registry    ports.Registry
	Discovery   ports.Discovery
	NATSClient  ports.NATSClient
	Metrics     ports.MetricsReader
	RuntimeView *coreruntime.View
	P99Window   string
}

func SetupDependencies(resources Resources) *Container {
	container := &Container{
		Store:       resources.Store,
		FlowManager: resources.FlowManager,
		Registry:    resources.Registry,
		Discovery:   resources.Discovery,
		NATSClient:  resources.NATSClient,
		Metrics:     resources.Metrics,
		RuntimeView: resources.RuntimeView,
		P99Window:   resources.P99Window,
		CDCService:  drivergrpc.NewCDCService(resources.Store, resources.FlowManager, resources.Registry, resources.Discovery, resources.NATSClient, resources.RuntimeView, resources.Metrics, resources.P99Window),
	}
	GlobalContainer = container
	return container
}
