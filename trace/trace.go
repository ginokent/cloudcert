package trace

// NOTE: OpenTelemetry in Go https://zenn.dev/ww24/articles/beae98be198c94

import (
	"context"
	"os"
	"sync"

	gcloudtrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/newtstat/cloudacme/config"
	"github.com/rec-logger/rec.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/xerrors"
)

// NewExporter TODO
// nolint: ireturn
func NewExporter() (exporter sdktrace.SpanExporter, err error) {
	switch config.SpanExporter() {
	case "gcloud":
		exporter, err = gcloudtrace.New(gcloudtrace.WithProjectID(config.GoogleCloudProject()))
		if err != nil {
			return nil, xerrors.Errorf("gcloudtrace.New: %w", err)
		}
	case "stdout":
		exporter, err = stdouttrace.New(stdouttrace.WithWriter(os.Stdout))
		if err != nil {
			return nil, xerrors.Errorf("stdouttrace.New: %w", err)
		}
	default:
		exporter = nil
	}

	return exporter, nil
}

// NewResource TODO.
func NewResource(serviceName, version string) *resource.Resource {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(version),
		semconv.TelemetrySDKLanguageGo,
	}

	return resource.NewWithAttributes(semconv.SchemaURL, attrs...)
}

var (
	tracer      trace.Tracer    // nolint: gochecknoglobals
	tracerMutex = &sync.Mutex{} // nolint: gochecknoglobals
)

var (
	_noopTracer    trace.Tracer   // nolint: gochecknoglobals
	noopTracerOnce = &sync.Once{} // nolint: gochecknoglobals
)

// cf. https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/github.com/gorilla/mux/otelmux/example/server.go
func SetupTracerProvider(l *rec.Logger) (shutdown func(), err error) {
	l.Info("🔔 start OpenTelemetry Tracer Provider")

	exporter, err := NewExporter()
	if err != nil {
		return nil, xerrors.Errorf("NewExporter: %w", err)
	}

	if exporter == nil {
		return func() { /* noop */ l.Info("🔕 shutdown OpenTelemetry Tracer Provider") }, nil
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(NewResource(config.AppName, config.Version())),
	)
	otel.SetTracerProvider(tracerProvider)
	tracerMutex.Lock()
	tracer = tracerProvider.Tracer(config.TracerName)
	tracerMutex.Unlock()

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})) // NOTE: TextMapPropagator について調べる

	shutdown = func() {
		defer l.Info("🔕 shutdown OpenTelemetry Tracer Provider")

		flushCtx, flushCancel := context.WithTimeout(context.Background(), config.ShutdownTimeout())
		defer flushCancel()

		if err := tracerProvider.ForceFlush(flushCtx); err != nil {
			rec.L().F().Errorf("failed to flush tracer provider: %v", err)
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.ShutdownTimeout())
		defer shutdownCancel()

		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			rec.L().F().Errorf("failed to shutdown tracer provider: %v", err)
		}
	}

	return shutdown, nil
}

func noopTracer() trace.Tracer { // nolint: ireturn
	noopTracerOnce.Do(func() {
		_noopTracer = trace.NewNoopTracerProvider().Tracer(config.TracerName)
	})
	return _noopTracer
}

func Start(parent context.Context, spanName string, opts ...trace.SpanStartOption) (child context.Context, span trace.Span) { // nolint: ireturn
	if tracer == nil {
		return noopTracer().Start(parent, spanName, opts...)
	}
	return tracer.Start(parent, spanName, opts...)
}

func StartFunc(parent context.Context, spanName string, opts ...trace.SpanStartOption) func(spanFunction func(child context.Context) (err error)) error {
	return func(spanFunction func(context.Context) error) error {
		child, span := Start(parent, spanName, opts...)
		defer span.End()

		if err := spanFunction(child); err != nil {
			return err
		}

		return nil
	}
}
