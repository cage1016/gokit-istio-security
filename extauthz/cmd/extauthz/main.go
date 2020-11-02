package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"

	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/qeek-dev/ext-authz/internal/app/extauthz"
)

const (
	defServiceName = "extauthz"
	defLogLevel    = "error"
	defServiceHost = "localhost"
	defGRPCPort    = "50051"
	defAuthzsvcURL = "localhost:10121"
	defByPass      = "false"

	envServiceName = "QS_SERVICE_NAME"
	envLogLevel    = "QS_LOG_LEVEL"
	envServiceHost = "QS_SERVICE_HOST"
	envGRPCPort    = "QS_GRPC_PORT"
	envAuthzsvcURL = "QS_AUTHZ_URL"
	envByPass      = "QS_BY_PASS"
)

type config struct {
	serviceName string
	logLevel    string
	serviceHost string
	grpcPort    string
	authzSvcURL string
	byPass      bool
}

// Env reads specified environment variable. If no value has been found,
// fallback is returned.
func env(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = level.NewFilter(logger, level.AllowInfo())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	cfg := loadConfig(logger)
	logger = log.With(logger, "service", cfg.serviceName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// addsvc grpc connection
	var conn *grpc.ClientConn
	{
		var err error
		if cfg.authzSvcURL != "" {
			conn, err = grpc.Dial(cfg.authzSvcURL, grpc.WithInsecure())
			if err != nil {
				level.Error(logger).Log("serviceName", cfg.authzSvcURL, "error", err)
				os.Exit(1)
			}
		}
	}

	svc := extauthz.NewAuthorizationServer(conn, cfg.byPass, logger)

	hs := health.NewServer()
	hs.SetServingStatus(cfg.serviceName, healthgrpc.HealthCheckResponse_SERVING)

	wg := &sync.WaitGroup{}

	go startGRPCServer(ctx, wg, svc, cfg.grpcPort, hs, logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()
	wg.Wait()

	fmt.Println("main: all goroutines have told us they've finished")
}

func loadConfig(logger log.Logger) (cfg config) {
	cfg.serviceName = env(envServiceName, defServiceName)
	cfg.logLevel = env(envLogLevel, defLogLevel)
	cfg.serviceHost = env(envServiceHost, defServiceHost)
	cfg.grpcPort = env(envGRPCPort, defGRPCPort)
	cfg.authzSvcURL = env(envAuthzsvcURL, defAuthzsvcURL)
	b, err := strconv.ParseBool(env(envByPass, defByPass))
	if err != nil {
		level.Error(logger).Log("loadConfig", "fail", "err", err)
		cfg.byPass = false
	} else {
		cfg.byPass = b
	}
	return cfg
}

func startGRPCServer(ctx context.Context, wg *sync.WaitGroup, srv extauthz.AuthorizationServer, port string, hs *health.Server, logger log.Logger) {
	wg.Add(1)
	defer wg.Done()

	p := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", p)
	if err != nil {
		level.Error(logger).Log("protocol", "GRPC", "listen", port, "err", err)
		os.Exit(1)
	}

	var server *grpc.Server
	level.Info(logger).Log("protocol", "GRPC", "exposed", port)
	server = grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	auth.RegisterAuthorizationServer(server, &srv)
	healthgrpc.RegisterHealthServer(server, hs)
	reflection.Register(server)

	go func() {
		// service connections
		err = server.Serve(listener)
		if err != nil {
			fmt.Printf("grpc serve : %s\n", err)
		}
	}()

	<-ctx.Done()

	// ignore error since it will be "Err shutting down server : context canceled"
	server.GracefulStop()

	fmt.Println("grpc server gracefully stopped")
}
