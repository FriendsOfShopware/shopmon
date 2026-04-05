// Package telemetry sets up OpenTelemetry tracing and logging.
package telemetry

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Setup initializes OpenTelemetry tracing and logging with OTLP HTTP exporters.
// It sets slog's default logger to a handler that sends logs via OTLP and also
// writes to stderr. Returns a shutdown function that should be called on application exit.
// If endpoint is empty, telemetry is disabled and a no-op shutdown is returned.
func Setup(ctx context.Context, serviceName, traceEndpoint, logEndpoint string) (shutdown func(context.Context) error) {
	if traceEndpoint == "" && logEndpoint == "" {
		return func(context.Context) error { return nil }
	}

	// Use a separate context with timeout so exporter creation doesn't block startup.
	setupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := resource.New(setupCtx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		slog.Error("failed to create OTel resource", "error", err)
		return func(context.Context) error { return nil }
	}

	var shutdownFuncs []func(context.Context) error

	// Tracing
	if traceEndpoint != "" {
		traceExporter, err := otlptracehttp.New(setupCtx,
			otlptracehttp.WithEndpointURL(traceEndpoint),
		)
		if err != nil {
			slog.Error("failed to create OTLP trace exporter", "error", err)
		} else {
			tp := sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(traceExporter),
				sdktrace.WithResource(res),
			)
			otel.SetTracerProvider(tp)
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			))
			shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
			slog.Info("OpenTelemetry tracing enabled", "endpoint", traceEndpoint, "service", serviceName)
		}
	}

	// Logging
	if logEndpoint != "" {
		logExporter, err := otlploghttp.New(setupCtx,
			otlploghttp.WithEndpointURL(ensurePath(logEndpoint, "/v1/logs")),
		)
		if err != nil {
			slog.Error("failed to create OTLP log exporter", "error", err)
		} else {
			lp := log.NewLoggerProvider(
				log.WithProcessor(log.NewBatchProcessor(logExporter)),
				log.WithResource(res),
			)
			otelHandler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(lp))
			stderrHandler := slog.NewTextHandler(os.Stderr, nil)
			slog.SetDefault(slog.New(newMultiHandler(stderrHandler, otelHandler)))
			shutdownFuncs = append(shutdownFuncs, lp.Shutdown)
			slog.Info("OpenTelemetry logging enabled", "endpoint", logEndpoint, "service", serviceName)
		}
	}

	return func(ctx context.Context) error {
		var firstErr error
		for _, fn := range shutdownFuncs {
			if err := fn(ctx); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
}

// ensurePath appends defaultPath to the endpoint URL if it has no path set.
func ensurePath(endpoint, defaultPath string) string {
	u, err := url.Parse(endpoint)
	if err != nil || (u.Path != "" && u.Path != "/") {
		return endpoint
	}
	u.Path = defaultPath
	return u.String()
}

// multiHandler fans out log records to multiple slog handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return newMultiHandler(handlers...)
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return newMultiHandler(handlers...)
}
