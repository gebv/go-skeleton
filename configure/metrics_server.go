package configure

import (
	"context"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type logFunc func(v ...interface{})

func (l logFunc) Println(v ...interface{}) {
	l(v...)
}

func RunDebugMux(ctx context.Context, address string) {
	l := zap.L().Named("debugMux")
	sugar := l.Sugar()

	http.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		ErrorLog:      logFunc(sugar.Warn),
		ErrorHandling: promhttp.HTTPErrorOnError,
	}))

	l.Info("Starting server...", zap.String("address", address))
	lis, err := net.Listen("tcp", address)
	if err != nil {
		l.Panic("Failed to listen.", zap.String("address", address), zap.Error(err))
	}
	l.Info("Metrics server for prometheus run", zap.String("address", address))

	s := &http.Server{}
	go func() {
		if err := s.Serve(lis); err != nil && err != http.ErrServerClosed {
			l.Error("Serve error.", zap.Error(err))
		}
	}()

	<-ctx.Done()
	if err := s.Close(); err != nil {
		l.Error("Close error.", zap.Error(err))
	} else {
		l.Info("Server stopped.")
	}
}
