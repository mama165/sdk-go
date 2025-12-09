package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UnaryLoggingInterceptor is a gRPC interceptor that logs the request
// Reads and logs the request of incoming HTTP requests
// Useful for debugging gRPC binary during development or in controlled environments.
// ⚠️ Note: The request is logged in plain text.
// ⚠️ Note: Do not use this in production without filtering or masking sensitive fields
// To secure : passwords, tokens, or credentials.
func UnaryLoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		_, ok := req.(*emptypb.Empty)
		if ok {
			logger.Debug("[gRPC] incoming request",
				slog.String("method", info.FullMethod),
				slog.String("request", "the body is empty"),
			)
		} else {
			logger.Debug("[gRPC] incoming request",
				slog.String("method", info.FullMethod),
				slog.Any("request", req),
			)
		}
		return handler(ctx, req)
	}
}
