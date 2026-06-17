package platform

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterHealth attaches the standard gRPC health service to a server.
func RegisterHealth(server *grpc.Server, serviceName string) {
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus(serviceName, healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
}

// RunGRPC serves gRPC until SIGINT/SIGTERM, then calls GracefulStop.
func RunGRPC(lis net.Listener, server *grpc.Server, logger util.Logger, name string) error {
	if name == "" {
		name = "grpc"
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Transport().Info().Str("service", name).Str("addr", lis.Addr().String()).Msg("grpc server listening")
		if err := server.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("%s server: %w", name, err)
	case sig := <-sigCh:
		logger.Service().Info().Str("service", name).Str("signal", sig.String()).Msg("shutdown initiated")
	}

	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(30 * time.Second):
		server.Stop()
	}

	logger.Service().Info().Str("service", name).Msg("grpc server stopped")
	return nil
}
