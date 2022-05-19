package entrypoint

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/newtstat/cloudacme/controller/router"
	"github.com/rec-logger/rec.go" // NOTE: github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/google.golang.org/grpc/otelgrpc
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func StartGRPCServerWithGatewayAsync(ctx context.Context, address string, shutdownTimeout time.Duration, l *rec.Logger) (shutdown func(), errChan <-chan error) {
	errCh := make(chan error, 1)

	grpcGatewayRouter, err := router.NewGRPCGatewayRouter(ctx, address, l)
	if err != nil {
		errCh <- errors.Errorf("router.NewGRPCGatewayRouter: %w", err)
		return func() {}, errCh
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcGatewayRouter)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		errCh <- errors.Errorf("net.Listen: %w", err)
		return func() {}, errCh
	}

	grpcServer := router.NewGRPCServer(l)

	server := &http.Server{
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// NOTE: cf. https://github.com/grpc/grpc-go/issues/555#issuecomment-443293451
			// NOTE: cf. https://github.com/philips/grpc-gateway-example/issues/22#issuecomment-490733965
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
			} else {
				mux.ServeHTTP(w, r)
			}
		}), &http2.Server{}),
	}

	go func() {
		l.F().Infof("🔊 start gRPC Server with gRPC-Gateway: %s", listener.Addr().String())
		defer l.F().Infof("🔇 shutdown gRPC Server and gRPC-Gateway: %s", listener.Addr().String())

		if err := server.Serve(listener); err != nil {
			errCh <- errors.Errorf("http.Serve: %w", err)
			return
		}
		errCh <- nil
		return // nolint: gosimple
	}()

	shutdown = func() {
		grpcServer.GracefulStop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			l.E().Error(errors.Errorf("server.Shutdown: %w", err))
			return
		}
	}

	return shutdown, errCh
}
