package grpc

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/mama165/sdk-go/logs"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestUnaryLoggingInterceptorWithEmptyRequest(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)

	req := &emptypb.Empty{}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/EmptyCall"}
	fakeHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "OK", nil
	}
	interceptor := UnaryLoggingInterceptor(logger)

	_, err := interceptor(ctx, req, info, fakeHandler)
	require.NoError(t, err)

	logOutput := buf.String()
	require.Contains(t, logOutput, "the body is empty")
	require.Contains(t, logOutput, info.FullMethod)
}

func TestUnaryLoggingInterceptorWithNonEmptyRequest(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	logger := logs.GetLoggerFromBufferWithLogger(&buf, slog.LevelDebug)
	req := struct {
		Message string
	}{"hello"}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/NonEmptyCall"}
	fakeHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "OK", nil
	}

	interceptor := UnaryLoggingInterceptor(logger)
	_, err := interceptor(ctx, req, info, fakeHandler)
	require.NoError(t, err)

	logOutput := buf.String()
	require.NotContains(t, logOutput, "the body is empty")
	require.Contains(t, logOutput, info.FullMethod)
}
