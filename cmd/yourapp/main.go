package main

import (
	"context"
	"flag"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/postgresql"

	"github.com/gebv/go-skeleton/api/yourapp_api"
	"github.com/gebv/go-skeleton/configure"
	loggerInterceptor "github.com/gebv/go-skeleton/interceptors/logger"
	recoverInterceptor "github.com/gebv/go-skeleton/interceptors/recover"
	settingsInterceptor "github.com/gebv/go-skeleton/interceptors/settings"
	"github.com/gebv/go-skeleton/logger"
	settingsService "github.com/gebv/go-skeleton/settings"
)

var (
	VERSION          = "dev"
	consulAddressF   = flag.String("consul-address", "127.0.0.1:8500", "Consul address (host and port).")
	consulConfigKeyF = flag.String("consul-config-key", "consul-config-key", "Consul config key.")
)

func main() {
	rand.Seed(time.Now().UnixNano())
	logger.Configure("INFO")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	zap.L().Info("Starting...", zap.String("version", VERSION))
	defer func() { zap.L().Info("Done.") }()

	// settings of service
	reloader, consulClient := settingsService.ConnectAndRunReloader(ctx, *consulAddressF, *consulConfigKeyF)
	settings := reloader.Settings()

	// register service to in Consul
	err := configure.RegisterService(consulClient, settings.Prometheus.RegSerivceName, settings.Prometheus.MetricsListenAddress)
	if err != nil {
		zap.L().Panic("Failed to register service in Consul.", zap.Error(err))
	}

	handleTerm(cancel)

	// Postgres
	sqlDB := configure.SetupPostgres(settings)
	reformDB := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(zap.L().Sugar().Debugf))

	// gRPC listener
	lis, err := net.Listen("tcp", settings.API.GRPCListenAddress)
	if err != nil {
		zap.L().Panic("gRPC failed to listen.", zap.Error(err), zap.String("address", settings.API.GRPCListenAddress))
	}

	// TODO: init services
	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			settingsInterceptor.Unary(reloader, VERSION),
			recoverInterceptor.Unary(),
			// TODO: auth interceptor
			loggerInterceptor.Unary(),
		)),
		grpc.StreamInterceptor(middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor,
			settingsInterceptor.Stream(reloader),
			recoverInterceptor.Stream(),
			// TODO: auth interceptor
			loggerInterceptor.Stream(),
		)),
	)

	// graceful stop takes up to stopTimeout
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		go func() {
			<-ctx.Done()
			// TODO: stop services
			s.Stop()
			lis.Close()
		}()

		// TODO: graceful stop of services
		s.GracefulStop()
	}()

	if settings.Debug.EnableReflection {
		// для того что бы работа grpcurl (для отладки)
		reflection.Register(s)
	}

	// TODO: run services
	yourapp_api.RegisterUsersServer(s, nil)

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(s)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		// metrics of app for prometheus scraper
		configure.RunDebugMux(ctx, settings.Prometheus.MetricsListenAddress)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// grpc web server
		configure.RunGrpcWebServer(ctx, s, settings.API.GRPCWebListenAddress, []string{})
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// run grpc сервера
		if err := s.Serve(lis); err != nil {
			zap.L().Error("Failed run gRPC server.", zap.Error(err))
		}
		wg.Done()
	}()
	zap.L().Info("gRPC listen address", zap.String("address", lis.Addr().String()))

	zap.L().Info("Application is ready")
	wg.Wait()
	zap.L().Info("Bye bye")
}

func handleTerm(cancel context.CancelFunc) {
	// handle termination signals: first one gracefully, force exit on the second one
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGTERM, unix.SIGINT)
	go func() {
		s := <-signals
		zap.L().Warn("Shutting down.", zap.String("signal", unix.SignalName(s.(unix.Signal))))
		cancel()

		s = <-signals
		zap.L().Panic("Exiting!", zap.String("signal", unix.SignalName(s.(unix.Signal))))
	}()
}
