package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/newtstat/cloudacme/config"
	"github.com/newtstat/cloudacme/entrypoint"
	"github.com/newtstat/cloudacme/trace"
	"github.com/rec-logger/rec.go" // NOTE: github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/google.golang.org/grpc/otelgrpc
	"golang.org/x/xerrors"
)

func main() {
	l := rec.Must(rec.New(os.Stdout))
	rec.ReplaceDefaultLogger(l)

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	if err := run(shutdownChan, l); err != nil {
		l.E().Error(err)
		os.Exit(1)
	}
}

func run(shutdownChan <-chan os.Signal, l *rec.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l.F().Infof("🔆 start %s: GoVersion=%s Version=%s Revision=%s Timestamp=%s", config.AppName, config.GoVersion(), config.Version(), config.Revision(), config.Timestamp())
	defer l.F().Infof("💤 shutdown %s: GoVersion=%s Version=%s Revision=%s Timestamp=%s", config.AppName, config.GoVersion(), config.Version(), config.Revision(), config.Timestamp())

	cleanup, err := trace.SetupTracerProvider(l)
	if err != nil {
		return xerrors.Errorf("trace.SetupTracerProvider: %w", err)
	}
	defer cleanup()

	address := fmt.Sprintf("%s:%d", config.Addr(), config.Port())

	shutdown, errCh := entrypoint.StartGRPCServerWithGatewayAsync(ctx, address, config.ShutdownTimeout(), l)

	select {
	case <-ctx.Done():
		l.E().Info(ctx.Err())
	case sig := <-shutdownChan:
		l.F().Infof("catch the signal: %s", sig)
	case err := <-errCh:
		if err != nil {
			return xerrors.Errorf("entrypoint.StartGRPCGatewayAsync: %w", err)
		}
	}

	shutdown()

	time.Sleep(1 * time.Millisecond) // NOTE: wait shutdown log

	return nil
}
