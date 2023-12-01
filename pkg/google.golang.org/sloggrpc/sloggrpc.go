package sloggrpc

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
)

// interceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
//
// Copied from https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/logging/examples/slog/example_test.go
func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func UnaryDialOption(logger *slog.Logger) grpc.DialOption {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.FinishCall),
	}
	return grpc.WithUnaryInterceptor(logging.UnaryClientInterceptor(interceptorLogger(logger), opts...))
}

func StreamDialOption(logger *slog.Logger) grpc.DialOption {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.FinishCall),
	}
	return grpc.WithStreamInterceptor(logging.StreamClientInterceptor(interceptorLogger(logger), opts...))
}
