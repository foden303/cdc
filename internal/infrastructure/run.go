package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	drivendiscovery "github.com/foden/cdc/internal/adapters/driven/discovery"
	drivenmetrics "github.com/foden/cdc/internal/adapters/driven/metrics"
	"github.com/foden/cdc/internal/core/flow"
	coreruntime "github.com/foden/cdc/internal/core/runtime"
	"github.com/foden/cdc/internal/di"
)

func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	SetupLogger(cfg.Log)
	slog.Info("cdc starting",
		"grpc_port", cfg.Server.GRPCPort,
		"http_port", cfg.Server.HTTPPort,
		"nats_url", cfg.NATS.URL,
	)

	setupCtx, setupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer setupCancel()

	natsClient, err := SetupNATS(setupCtx, cfg.NATS)
	if err != nil {
		return err
	}

	store, err := SetupStorage(setupCtx, natsClient)
	if err != nil {
		natsClient.Close()
		return err
	}

	reg := SetupRegistry()
	poolManager := flow.NewPoolManager()
	runtimeRegistry := coreruntime.NewRegistry()
	runtimeMetrics := coreruntime.NewMetrics()
	runtimeView := coreruntime.NewView(runtimeRegistry, runtimeMetrics, flow.NewRuntimePoolMetricsProvider(poolManager))
	coreruntime.SetDefaults(runtimeRegistry, runtimeMetrics, runtimeView)
	disc := drivendiscovery.NewService()
	metricsReader := drivenmetrics.NewPrometheusClient(cfg.Prometheus.URL)
	flowManager := flow.NewManager(
		store,
		poolManager,
		reg,
		natsClient,
		disc,
		flow.WithMaxDeliver(cfg.NATS.MaxDeliver),
		flow.WithRuntime(runtimeRegistry, runtimeMetrics, runtimeView),
	)

	restoreCtx, restoreCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer restoreCancel()
	if err := flowManager.RestoreFlows(restoreCtx); err != nil {
		slog.Warn("failed to restore flows", "err", err)
	}

	container := di.SetupDependencies(di.Resources{
		Store:       store,
		FlowManager: flowManager,
		Registry:    reg,
		Discovery:   disc,
		NATSClient:  natsClient,
		Metrics:     metricsReader,
		RuntimeView: runtimeView,
		P99Window:   cfg.Prometheus.QueryWindow,
	})

	appServer := NewAppServer(cfg.Server, container)
	if err := appServer.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	slog.Info("received shutdown signal", "signal", sig)
	appServer.Stop()
	return nil
}
