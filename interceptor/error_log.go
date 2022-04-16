package interceptor

import (
	"context"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/rec-logger/rec.go"
	"google.golang.org/grpc"
)

func ErrorLogInterceptor(original *rec.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			l := contexts.GetLogger(ctx)
			l.With(rec.ErrorStacktrace(err)).F().Errorf("error=%v", err)
			return nil, err
		}

		return resp, nil
	}
}
