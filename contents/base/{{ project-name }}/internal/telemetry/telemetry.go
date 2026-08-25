// Package telemetry wires OpenTelemetry tracing fail-open: spans export over OTLP iff
// OTEL_EXPORTER_OTLP_ENDPOINT is set; otherwise tracing stays a no-op. The service name comes
// from OTEL_SERVICE_NAME, defaulting to the project name.
package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Init configures the global tracer provider and returns a shutdown func (never nil). It never
// fails the service: a missing endpoint means tracing is off, an exporter error is logged and
// ignored.
func Init(ctx context.Context, defaultServiceName string) func(context.Context) {
	noop := func(context.Context) {}

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return noop
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	// The exporter reads the OTEL_EXPORTER_OTLP_* env vars itself.
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		slog.Warn("otel exporter init failed; tracing disabled", "error", err)
		return noop
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)

	return func(ctx context.Context) {
		if err := tp.Shutdown(ctx); err != nil {
			slog.Warn("otel shutdown failed", "error", err)
		}
	}
}
