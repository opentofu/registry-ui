// Package telemetry sets up OpenTelemetry tracing and logging for the application.
package telemetry

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/go-logr/stdr"
	"github.com/lmittmann/tint"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/encoding/gzip"

	"github.com/opentofu/registry-ui/pkg/config"
)

const (
	DefaultServiceName = "opentofu-registry-backend"

	// BatchJobIDKey is the attribute key for batch job correlation
	BatchJobIDKey = "batch.job.id"
	// BatchJobNameKey is the attribute key for batch job name
	BatchJobNameKey = "batch.job.name"
)

func SetupTelemetry(ctx context.Context, config config.TelemetryConfig) (context.Context, func(), error) {
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
		return ctx, func() {}, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP grpc exporter using config
	// Use gzip compression to reduce payload size and avoid gRPC message size limits
	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithHeaders(config.Headers),
		otlptracegrpc.WithCompressor(gzip.Name),
	}

	// Add insecure option if configured
	if config.Insecure {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return ctx, func() {}, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Set up tracer provider with span limits to prevent oversized spans
	// AttributeValueLengthLimit truncates large attribute values (e.g., stderr output)
	// Combined with gzip compression, this prevents gRPC message size limit issues
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBlocking(),
			sdktrace.WithMaxExportBatchSize(128),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(otelResource),
		sdktrace.WithRawSpanLimits(sdktrace.SpanLimits{
			AttributeValueLengthLimit: 4096, // Truncate attribute values over 4KB
			AttributeCountLimit:       128,  // Max 128 attributes per span
			EventCountLimit:           128,  // Max 128 events per span
			LinkCountLimit:            128,  // Max 128 links per span
		}),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set up logger provider (for otelslog) if logging is enabled
	var loggerProvider *sdklog.LoggerProvider
	if config.Logging {
		logExporterOpts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(config.Endpoint),
			otlploggrpc.WithHeaders(config.Headers),
			otlploggrpc.WithCompressor(gzip.Name),
		}
		if config.Insecure {
			logExporterOpts = append(logExporterOpts, otlploggrpc.WithInsecure())
		}

		logExporter, err := otlploggrpc.New(ctx, logExporterOpts...)
		if err != nil {
			return ctx, func() {}, fmt.Errorf("failed to create OTLP log exporter: %w", err)
		}

		loggerProvider = sdklog.NewLoggerProvider(
			sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
			sdklog.WithResource(otelResource),
		)
		global.SetLoggerProvider(loggerProvider)
	}

	prop := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(prop)

	logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile))
	otel.SetLogger(logger)

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.ErrorContext(ctx, "OpenTelemetry error", "error", err)
	}))

	shutdown := func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown tracer provider", "error", err)
		}
		if loggerProvider != nil {
			if err := loggerProvider.Shutdown(ctx); err != nil {
				slog.Error("Failed to shutdown logger provider", "error", err)
			}
		}
	}

	// Instrument all HTTP requests globally
	http.DefaultTransport = otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Host)
		}),
	)

	return ctx, shutdown, nil
}

// Tracer is intended to be the primary way of creating a tracer in this application,
// It will attempt to set the name of the tracer to that of the function calling this method.
// Please ensure that your methods that call this are nicely named!
// ---
// Based on the OpenTofu implementation here: https://github.com/opentofu/opentofu/blame/481798ab36e56c4d8109377ebacc401a9ad82974/internal/tracing/utils.go#L25
func Tracer() trace.Tracer {
	pc, _, _, ok := runtime.Caller(1)
	if !ok || runtime.FuncForPC(pc) == nil {
		return otel.Tracer("")
	}

	// We use the import path of the caller function as the tracer name.
	return otel.GetTracerProvider().Tracer(extractImportPath(runtime.FuncForPC(pc).Name()))
}

// extractImportPath extracts the import path from a full function name.
// the function names returned by runtime.FuncForPC(pc).Name() can be in the following formats
//
//	main.(*MyType).MyMethod
//	github.com/you/pkg.(*SomeType).Method-fm
//	github.com/you/pkg.functionName
func extractImportPath(fullName string) string {
	lastSlash := strings.LastIndex(fullName, "/")
	if lastSlash == -1 {
		// When there is no slash, then use everything before the first dot
		if dot := strings.Index(fullName, "."); dot != -1 {
			return fullName[:dot]
		}
		log.Printf("unable to extract import path from function name: %q. Tracing may be incomplete.", fullName)
		return "unknown"
	}

	dotAfterSlash := strings.Index(fullName[lastSlash:], ".")
	if dotAfterSlash == -1 {
		log.Printf("[WARN] unable to extract import path from function name: %q. Tracing may be incomplete.", fullName)
		return "unknown"
	}

	return fullName[:lastSlash+dotAfterSlash]
}

func SetupLogger() {
	// Create tint handler for colored, readable console logs
	tintHandler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: "15:04:05",
	})

	// Wrap otelHandler with level filter to match tint handler
	baseOtelHandler := otelslog.NewHandler(DefaultServiceName)
	otelHandler := &levelFilterHandler{
		handler: baseOtelHandler,
		level:   slog.LevelInfo,
	}

	multi := multiHandler{[]slog.Handler{tintHandler, otelHandler}}

	logger := slog.New(multi)
	slog.SetDefault(logger)
}

// LinkedSpanStart starts a new root span that is linked to the current span in the context.
// This creates a separate trace while maintaining a reference to the parent span.
// Use this for batch processing where each item should have its own trace but be
// queryable via the batch.job.id attribute.
func LinkedSpanStart(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// Get the parent span context to create a link
	parentSpanCtx := trace.SpanContextFromContext(ctx)

	// Prepend WithNewRoot and WithLinks to the provided options
	baseOpts := []trace.SpanStartOption{
		trace.WithNewRoot(),
	}

	// Only add the link if there's a valid parent span
	if parentSpanCtx.IsValid() {
		baseOpts = append(baseOpts, trace.WithLinks(trace.Link{SpanContext: parentSpanCtx}))
	}

	// Append user-provided options (so they can override if needed)
	allOpts := append(baseOpts, opts...)

	return Tracer().Start(ctx, name, allOpts...)
}

// TestOTLPConnection tests connectivity to the OTLP endpoint by flushing traces and logs.
// Call this after SetupTelemetry to verify the endpoint is reachable.
func TestOTLPConnection(ctx context.Context, cfg config.TelemetryConfig) error {
	// Test trace connectivity
	tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	if !ok {
		return fmt.Errorf("tracer provider is not configured")
	}
	if err := tp.ForceFlush(ctx); err != nil {
		return fmt.Errorf("failed to flush traces to OTLP endpoint: %w", err)
	}
	slog.InfoContext(ctx, "OTLP trace connection test successful")

	// Test log connectivity if logging is enabled
	if cfg.Logging {
		lp, ok := global.GetLoggerProvider().(*sdklog.LoggerProvider)
		if !ok {
			return fmt.Errorf("logger provider is not configured")
		}
		if err := lp.ForceFlush(ctx); err != nil {
			return fmt.Errorf("failed to flush logs to OTLP endpoint: %w", err)
		}
		slog.InfoContext(ctx, "OTLP log connection test successful")
	}

	return nil
}
