package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/foden/cdc/config"
	drivergrpc "github.com/foden/cdc/internal/adapters/driver/grpc"
	"github.com/foden/cdc/internal/core/ports"
	"github.com/foden/cdc/internal/di"
)

// AppServer manages both gRPC and the HTTP REST/metrics server.
type AppServer struct {
	cfg         config.ServerConfig
	grpcServer  *grpc.Server
	httpServer  *http.Server
	flowManager ports.FlowManager
	natsClient  ports.NATSClient
	cdcService  *drivergrpc.CDCService
}

// NewAppServer creates a new combined gRPC + HTTP server.
func NewAppServer(
	cfg config.ServerConfig,
	container *di.Container,
) *AppServer {
	return &AppServer{
		cfg:         cfg,
		flowManager: container.FlowManager,
		natsClient:  container.NATSClient,
		cdcService:  container.CDCService,
	}
}

// Start launches the gRPC server and the HTTP server (metrics + health + grpc-gateway).
func (s *AppServer) Start() error {
	// 1. gRPC server with request_id interceptor
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(requestIDInterceptor()),
	)

	cdcpb.RegisterCDCServiceServer(s.grpcServer, s.cdcService)

	grpcAddr := fmt.Sprintf(":%d", s.cfg.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("gRPC listen on %s: %w", grpcAddr, err)
	}

	go func() {
		slog.Info("gRPC server started", "port", s.cfg.GRPCPort)
		if err := s.grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "err", err)
		}
	}()

	// 2. HTTP server: /metrics + /health + grpc-gateway (future)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	if err := cdcpb.RegisterCDCServiceHandlerServer(context.Background(), gwMux, s.cdcService); err != nil {
		return fmt.Errorf("register CDC REST gateway: %w", err)
	}
	mux.Handle("/", gwMux)

	handler := corsMiddleware(mux)

	httpAddr := fmt.Sprintf(":%d", s.cfg.HTTPPort)
	s.httpServer = &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		slog.Info("HTTP server started", "port", s.cfg.HTTPPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "err", err)
		}
	}()

	return nil
}

// Stop performs an ordered graceful shutdown with per-phase timeouts.
//
// Phase 1: Stop flow manager (stops all workers, drains events) — 10s timeout
// Phase 2: Close NATS connection
// Phase 3: Shutdown HTTP server — 5s timeout
// Phase 4: GracefulStop gRPC server
func (s *AppServer) Stop() {
	slog.Info("initiating ordered shutdown")

	// Phase 1: Stop flow manager (stops all workers, drains events)
	if s.flowManager != nil {
		slog.Info("shutdown phase 1: stopping flow manager")
		done := make(chan struct{})
		go func() {
			s.flowManager.Stop()
			close(done)
		}()

		select {
		case <-done:
			slog.Info("shutdown phase 1: flow manager stopped")
		case <-time.After(10 * time.Second):
			slog.Warn("shutdown phase 1: flow manager stop timed out after 10s")
		}
	}

	// Phase 2: Close NATS connection
	if s.natsClient != nil {
		slog.Info("shutdown phase 2: closing NATS connection")
		s.natsClient.Close()
		slog.Info("shutdown phase 2: NATS connection closed")
	}

	// Phase 3: Shutdown HTTP server
	if s.httpServer != nil {
		slog.Info("shutdown phase 3: shutting down HTTP server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			slog.Warn("shutdown phase 3: HTTP server shutdown error", "err", err)
		} else {
			slog.Info("shutdown phase 3: HTTP server stopped")
		}
	}

	// Phase 4: GracefulStop gRPC server
	if s.grpcServer != nil {
		slog.Info("shutdown phase 4: stopping gRPC server")
		s.grpcServer.GracefulStop()
		slog.Info("shutdown phase 4: gRPC server stopped")
	}

	slog.Info("shutdown complete")
}

// requestIDInterceptor injects a request_id into the context for every unary RPC.
func requestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		requestID := uuid.New().String()

		// Check if client sent a request_id in metadata
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-request-id"); len(vals) > 0 {
				requestID = vals[0]
			}
		}

		// Add request_id to outgoing metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)

		ctx = context.WithValue(ctx, requestIDKey{}, requestID)

		resp, err := handler(ctx, req)
		code := status.Code(err)
		attrs := []any{
			"request_id", requestID,
			"method", info.FullMethod,
			"code", code.String(),
			"duration_ms", time.Since(start).Milliseconds(),
		}
		if err != nil {
			slog.Warn("grpc request failed", append(attrs, "err", err)...)
		} else {
			slog.Info("grpc request completed", attrs...)
		}

		return resp, err
	}
}

// requestIDKey is the context key for storing request IDs.
type requestIDKey struct{}

// corsMiddleware adds CORS headers for frontend access.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
