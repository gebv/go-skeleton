package settings

import (
	"context"
	"runtime/pprof"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/gebv/go-skeleton/settings"
)

func Unary(reloader *settings.Reloader, appVersion string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		s := reloader.Settings()
		ctx = settings.Set(ctx, s)

		reqInfo, err := settings.ParseRequestMetaData(ctx)
		if err != nil {
			settings.GetLogger(ctx).Warn("Failed parse MD from request.", zap.Error(err))
		} else {
			ctx = settings.SetRequestMetaData(ctx, reqInfo)
		}

		l := zap.L().Named(info.FullMethod).With(
			zap.String("request_id", reqInfo.RequestID),
			zap.String("backend_version", appVersion),
		)

		if err := grpc.SetTrailer(ctx, metadata.Pairs("request-id", reqInfo.RequestID, "backend-version", appVersion)); err != nil {
			settings.GetLogger(ctx).Warn("Failed to send request-id trailer.", zap.Error(err))
		}

		// if s.Dev {
		// 	l = l.WithOptions(settings.LoggerWithLevel(zapcore.DebugLevel))
		// }
		ctx = settings.SetLogger(ctx, l)

		// add pprof labels for more useful profiles
		defer pprof.SetGoroutineLabels(ctx)
		ctx = pprof.WithLabels(ctx, pprof.Labels("method", info.FullMethod))
		pprof.SetGoroutineLabels(ctx)

		return handler(ctx, req)
	}
}

func Stream(reloader *settings.Reloader) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		s := reloader.Settings()
		ctx := settings.Set(ss.Context(), s)

		reqInfo, err := settings.ParseRequestMetaData(ctx)
		if err != nil {
			settings.GetLogger(ctx).Warn("Failed parse MD from request.", zap.Error(err))
		} else {
			ctx = settings.SetRequestMetaData(ctx, reqInfo)
		}

		l := zap.L().Named(info.FullMethod).With(zap.String("request_id", reqInfo.RequestID))

		// // if s.Dev {
		// // 	l = l.WithOptions(settings.LoggerWithLevel(zapcore.DebugLevel))
		// // }
		ctx = settings.SetLogger(ctx, l)

		// add pprof labels for more useful profiles
		defer pprof.SetGoroutineLabels(ctx)
		ctx = pprof.WithLabels(ctx, pprof.Labels("method", info.FullMethod))
		pprof.SetGoroutineLabels(ctx)

		wrapped := middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx
		return handler(srv, wrapped)
	}
}
