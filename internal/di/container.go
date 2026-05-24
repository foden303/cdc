package di

import (
	drivergrpc "github.com/foden/cdc/internal/adapters/driver/grpc"
	"github.com/foden/cdc/internal/core/ports"
	"github.com/foden/cdc/internal/core/service"
)

type Container struct {
	Store       ports.Store
	FlowManager ports.FlowManager
	Registry    ports.Registry
	Discovery   ports.Discovery
	NATSClient  ports.NATSClient

	CDCService       *drivergrpc.CDCService
	DashboardService *service.DashboardService
}

var GlobalContainer *Container
