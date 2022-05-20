package router

import (
	"context"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ginokent/cloudacme/controller"
	"github.com/ginokent/cloudacme/interceptor"
	"github.com/ginokent/cloudacme/middleware"
	"github.com/ginokent/cloudacme/proto/generated/go/v1/cloudacme"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gw_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/nitpickers/nits.go"
	"github.com/rec-logger/rec.go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func NewGRPCServer(l *rec.Logger) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				// NOTE: https://github.com/grpc-ecosystem/go-grpc-middleware
				otelgrpc.UnaryServerInterceptor(), // NOTE: OpenTelemetry for gRPC Gateway -> gRPC Server
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				interceptor.LoggerInterceptor(l),
				interceptor.AccessLogInterceptor(),
				interceptor.ErrorLogInterceptor(l),
				grpc_validator.UnaryServerInterceptor(),
			),
		),
	)

	// register servers
	cloudacme.RegisterTestAPIServer(grpcServer, &controller.TestAPIController{})
	cloudacme.RegisterCertificatesServer(grpcServer, &controller.CertificatesController{})

	return grpcServer
}

// NewGRPCGatewayRouter TODO
// cf. https://github.com/grpc-ecosystem/grpc-gateway
func NewGRPCGatewayRouter(ctx context.Context, grpcServerEndpoint string, l *rec.Logger) (http.Handler, error) {
	mux := gw_runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				otelgrpc.UnaryClientInterceptor(), // NOTE: OpenTelemetry for gRPC Gateway -> gRPC Server
			),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                1 * time.Minute,
			Timeout:             1 * time.Minute,
			PermitWithoutStream: true,
		}),
	}

	register := func(ctx context.Context, mux *gw_runtime.ServeMux, endpoint string, opts []grpc.DialOption, fs ...func(ctx context.Context, mux *gw_runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)) error {
		for _, f := range fs {
			if err := f(ctx, mux, endpoint, opts); err != nil {
				fv := reflect.ValueOf(f)
				return errors.Errorf("%s: %w", runtime.FuncForPC(fv.Pointer()).Name(), err)
			}
		}
		return nil
	}

	// register handlers
	if err := register(ctx, mux, grpcServerEndpoint, opts,
		cloudacme.RegisterTestAPIHandlerFromEndpoint,
		cloudacme.RegisterCertificatesHandlerFromEndpoint,
	); err != nil {
		return nil, errors.Errorf("cloudacme.RegisterTestAPIHandlerFromEndpoint: %w", err)
	}

	middlewares := nits.HTTP.AddMiddlewares(
		middleware.AccessLogMiddleware(),
		middleware.ContextLoggerRequestMiddleware(l),
		middleware.RequestBodyBufferMiddleware(),
	)

	// NOTE: OpenTelemetry for client -> gRPC Gateway
	otelHandler := otelhttp.NewHandler(
		middlewares(mux),
		"gRPC-Gateway",
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string { return operation + r.URL.Path }),
	)

	return otelHandler, nil
}
