package di

import (
	drivergrpc "github.com/foden/cdc/internal/adapters/driver/grpc"
	"github.com/foden/cdc/internal/core/ports"
	coreruntime "github.com/foden/cdc/internal/core/runtime"
)

type Container struct {
	Store       ports.Store
	FlowManager ports.FlowManager
	Registry    ports.Registry
	Discovery   ports.Discovery
	NATSClient  ports.NATSClient
	Metrics     ports.MetricsReader
	RuntimeView *coreruntime.View
	P99Window   string

	CDCService *drivergrpc.CDCService
}

var GlobalContainer *Container
