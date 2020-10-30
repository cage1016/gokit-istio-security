package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/cage1016/gokit-istio-security/internal/app/authz/endpoints"
	"github.com/cage1016/gokit-istio-security/internal/app/authz/engine/opa"
	"github.com/cage1016/gokit-istio-security/internal/app/authz/service"
	"github.com/cage1016/gokit-istio-security/internal/app/authz/storage/postgres"
	postgresV1 "github.com/cage1016/gokit-istio-security/internal/app/authz/storage/v1/postgres"
	transportsgrpc "github.com/cage1016/gokit-istio-security/internal/app/authz/transports/grpc"
	transportshttp "github.com/cage1016/gokit-istio-security/internal/app/authz/transports/http"
	"github.com/cage1016/gokit-istio-security/internal/pkg/logconv"
	pb "github.com/cage1016/gokit-istio-security/pb/authz"
)

const (
	defZipkinV2URL   = ""
	defServiceName   = "authz"
	defLogLevel      = "info"
	defServiceHost   = "localhost"
	defHTTPPort      = "8180"
	defGRPCPort      = "8181"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "postgres"
	defDBPass        = "password"
	defDBName        = "authz"
	defDBSSLMode     = "disable"
	defDBSSLCert     = ""
	defDBSSLKey      = ""
	defDBSSLRootCert = ""

	envZipkinV2URL   = "QS_ZIPKIN_V2_URL"
	envServiceName   = "QS_SERVICE_NAME"
	envLogLevel      = "QS_LOG_LEVEL"
	envServiceHost   = "QS_SERVICE_HOST"
	envHTTPPort      = "QS_HTTP_PORT"
	envGRPCPort      = "QS_GRPC_PORT"
	envDBHost        = "QS_DB_HOST"
	envDBPort        = "QS_DB_PORT"
	envDBUser        = "QS_DB_USER"
	envDBPass        = "QS_DB_PASS"
	envDBName        = "QS_DB"
	envDBSSLMode     = "QS_DB_SSL_MODE"
	envDBSSLCert     = "QS_DB_SSL_CERT"
	envDBSSLKey      = "QS_DB_SSL_KEY"
	envDBSSLRootCert = "QS_DB_SSL_ROOT_CERT"
)

type config struct {
	serviceName string
	logLevel    string
	serviceHost string
	httpPort    string
	grpcPort    string
	zipkinV2URL string
	dbConfig    postgres.Config
	secret      string
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
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	cfg := loadConfig(logger)
	logger = level.NewFilter(logger, logconv.Atol(cfg.logLevel))
	logger = log.With(logger, "service", cfg.serviceName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine, err := opa.New(ctx, logger)
	if err != nil {
		level.Error(logger).Log("could_not_initialize_engine", err)
	}

	store, err := postgresV1.New(ctx, cfg.dbConfig, logger)
	if err != nil {
		level.Error(logger).Log("database", err)
		os.Exit(1)
	}

	tracer := initOpentracing()
	zipkinTracer := initZipkin(cfg.serviceName, cfg.httpPort, cfg.zipkinV2URL, logger)
	policyRefresher, err := service.NewPolicyRefresher(ctx, logger, store, engine)
	if err != nil {
		level.Error(logger).Log("policyRefresher", err.Error())
	}
	svc := service.New(ctx, store, engine, policyRefresher, logger)
	eps := endpoints.New(svc, logger, tracer, zipkinTracer)

	hs := health.NewServer()
	hs.SetServingStatus(cfg.serviceName, healthgrpc.HealthCheckResponse_SERVING)

	wg := &sync.WaitGroup{}

	go startHTTPServer(ctx, wg, eps, tracer, zipkinTracer, cfg.httpPort, logger)
	go startGRPCServer(ctx, wg, eps, tracer, zipkinTracer, cfg.grpcPort, hs, logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()
	wg.Wait()

	fmt.Println("main: all goroutines have told us they've finished")
}

func loadConfig(logger log.Logger) (cfg config) {
	dbConfig := postgres.Config{
		Host:        env(envDBHost, defDBHost),
		Port:        env(envDBPort, defDBPort),
		User:        env(envDBUser, defDBUser),
		Pass:        env(envDBPass, defDBPass),
		Name:        env(envDBName, defDBName),
		SSLMode:     env(envDBSSLMode, defDBSSLMode),
		SSLCert:     env(envDBSSLCert, defDBSSLCert),
		SSLKey:      env(envDBSSLKey, defDBSSLKey),
		SSLRootCert: env(envDBSSLRootCert, defDBSSLRootCert),
	}

	cfg.dbConfig = dbConfig
	cfg.serviceName = env(envServiceName, defServiceName)
	cfg.logLevel = env(envLogLevel, defLogLevel)
	cfg.serviceHost = env(envServiceHost, defServiceHost)
	cfg.httpPort = env(envHTTPPort, defHTTPPort)
	cfg.grpcPort = env(envGRPCPort, defGRPCPort)
	cfg.zipkinV2URL = env(envZipkinV2URL, defZipkinV2URL)

	return cfg
}

func initOpentracing() stdopentracing.Tracer {
	return stdopentracing.GlobalTracer()
}

func initZipkin(serviceName, httpPort, zipkinV2URL string, logger log.Logger) (zipkinTracer *zipkin.Tracer) {
	var (
		err           error
		hostPort      = fmt.Sprintf("localhost:%s", httpPort)
		useNoopTracer = (zipkinV2URL == "")
		reporter      = zipkinhttp.NewReporter(zipkinV2URL)
	)
	zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
	zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer))
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
	if !useNoopTracer {
		logger.Log("tracer", "Zipkin", "type", "Native", "URL", zipkinV2URL)
	}

	return
}

func startHTTPServer(ctx context.Context, wg *sync.WaitGroup, endpoints endpoints.Endpoints, tracer stdopentracing.Tracer, zipkinTracer *zipkin.Tracer, port string, logger log.Logger) {
	wg.Add(1)
	defer wg.Done()

	if port == "" {
		level.Error(logger).Log("protocol", "HTTP", "exposed", port, "err", "port is not assigned exist")
		return
	}

	p := fmt.Sprintf(":%s", port)
	// create a server
	srv := &http.Server{Addr: p, Handler: transportshttp.NewHTTPHandler(endpoints, tracer, zipkinTracer, logger)}
	level.Info(logger).Log("protocol", "HTTP", "exposed", port)
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			level.Info(logger).Log("Listen", err)
		}
	}()

	<-ctx.Done()

	// shut down gracefully, but wait no longer than 5 seconds before halting
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ignore error since it will be "Err shutting down server : context canceled"
	srv.Shutdown(shutdownCtx)

	level.Info(logger).Log("protocol", "HTTP", "Shutdown", "http server gracefully stopped")
}

func startGRPCServer(ctx context.Context, wg *sync.WaitGroup, endpoints endpoints.Endpoints, tracer stdopentracing.Tracer, zipkinTracer *zipkin.Tracer, port string, hs *health.Server, logger log.Logger) {
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
	pb.RegisterAuthzServer(server, transportsgrpc.MakeGRPCServer(endpoints, tracer, zipkinTracer, logger))
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
