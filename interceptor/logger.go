package interceptor

import (
	"context"

	"github.com/rec-logger/rec.go"
	"google.golang.org/grpc"
)

func LoggerInterceptor(original *rec.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctxWithLogger := rec.ContextWithLogger(ctx,
			original.With(
				rec.String("method", info.FullMethod),
			),
		)

		resp, err := handler(ctxWithLogger, req)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}
