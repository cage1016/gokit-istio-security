package extauthz

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct {
	logger log.Logger
}

func (s HealthServer) String() string {
	return "HealthServer"
}

func NewHealthServer(logger log.Logger) HealthServer {
	return HealthServer{
		logger: logger,
	}
}
func (s HealthServer) Check(ctx context.Context, request *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

// rpc Watch(HealthCheckRequest) returns (stream HealthCheckResponse);
func (s HealthServer) Watch(request *health.HealthCheckRequest, stream health.Health_WatchServer) error {
	// Send healthy
	if err := stream.Send(&health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}); err != nil {
		level.Error(s.logger).Log("err", "failed to send response into gRPC stream")
	}

	return nil
}
