package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/go-logr/stdr"
	"github.com/lmittmann/tint"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"github.com/opentofu/registry-ui/pkg/config"
)

const (
	DefaultServiceName = "opentofu-registry-backend"
)

func initTelemetry(ctx context.Context, config config.TelemetryConfig) (context.Context, func(), error) {
	// Check if tracing is enabled in config
	if !config.Enabled {
		slog.InfoContext(ctx, "OpenTelemetry tracing is disabled")
		return ctx, func() {}, nil // Silent mode - no tracing
	}

	// Check if exporter is configured
	if config.Exporter == "" {
		slog.InfoContext(ctx, "No telemetry exporter configured")
		return ctx, func() {}, nil
	}

	slog.InfoContext(ctx, "OpenTelemetry tracing enabled", "exporter", config.Exporter)

	// Use service name from config or default
	serviceName := DefaultServiceName
	if config.ServiceName != "" {
		slog.InfoContext(ctx, "Using service name from config", "service_name", config.ServiceName)
		serviceName = config.ServiceName
	}

	otelResource, err := resource.New(context.Background(),
		resource.WithOS(),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("2.0.0"),
			semconv.TelemetrySDKName("opentelemetry"),
			semconv.TelemetrySDKLanguageGo,
			semconv.TelemetrySDKVersion(sdk.Version()),
		),
	)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter using config
	exporterOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Endpoint),
	}
	
	// Add insecure option if configured
	if config.Insecure {
		exporterOpts = append(exporterOpts, otlptracehttp.WithInsecure())
	}
	
	exporter, err := otlptracehttp.New(ctx, exporterOpts...)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBlocking()),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(otelResource),
	)
	otel.SetTracerProvider(provider)

	prop := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(prop)

	logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile))
	otel.SetLogger(logger)

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.ErrorContext(ctx, "OpenTelemetry error", "error", err)
	}))

	shutdown := func() {
		if err := provider.Shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown tracer provider", "error", err)
		}
	}

	return ctx, shutdown, nil
}

func setupLogger() {
	// Create tint handler for colored, readable logs
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: "15:04:05",
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
