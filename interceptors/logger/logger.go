package logger

import (
	"context"

	"github.com/gebv/go-skeleton/settings"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		ctxLogger := settings.GetLogger(ctx)

		defer func() {
			if err != nil {
				ctxLogger.Info("Request details", zap.Error(err))
			} else {
				ctxLogger.Info("Request details")
			}
		}()

		res, err = handler(ctx, req)
		return
	}
}

func Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctxLogger := settings.GetLogger(ss.Context())

		defer func() {
			if err != nil {
				ctxLogger.Info("Stream request details", zap.Error(err))
			} else {
				ctxLogger.Info("Stream request details")
			}

		}()
		err = handler(srv, ss)
		return
	}
}
