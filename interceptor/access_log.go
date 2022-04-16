package interceptor

import (
	"context"
	"path"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/rec-logger/rec.go"
	"google.golang.org/grpc"
)

func AccessLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		l := contexts.GetLogger(ctx)
		l = l.With(
			rec.String("grpc.service", path.Dir(info.FullMethod)[1:]),
			rec.String("grpc.method", path.Base(info.FullMethod)),
			// rec.Sprintf("grpc.request", "%v", req),
		)

		defer l.F().Infof("access: error=%v <- grpc.method=%s", err, info.FullMethod)

		resp, err = handler(ctx, req)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}
