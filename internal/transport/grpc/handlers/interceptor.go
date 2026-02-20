package handlers

import (
	"context"
	"log/slog"

	"github.com/tainj/distributed_calculator2/pkg/logger"
	"google.golang.org/grpc"
)

func ContextWithLogger(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l.Info(ctx, "request started", slog.String("method", info.FullMethod))
		return handler(ctx, req)
	}
}
