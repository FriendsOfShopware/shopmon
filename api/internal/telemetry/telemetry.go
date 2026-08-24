// Package telemetry sets up OpenTelemetry tracing, metrics, and logging.
package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/metrics"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config describes the exporters and the resource attributes they report.
type Config struct {
	ServiceName   string
	Version       string
	DeploymentEnv string
	TraceEndpoint string
	LogEndpoint   string
	// MetricEndpoint is the OTLP HTTP metrics URL. Empty disables metrics
	// (global MeterProvider stays a no-op).
	MetricEndpoint string
	// SamplerRatio is the head sampling ratio in [0, 1]; 1 samples everything.
	SamplerRatio float64
}

// Setup initializes OpenTelemetry tracing, metrics, and logging with OTLP HTTP
// exporters. It sets slog's default logger to a handler that sends logs via
// OTLP and also writes to stderr. Returns a shutdown function that should be
// called on application exit. If all endpoints are empty, telemetry is disabled
// and a no-op shutdown is returned.
func Setup(ctx context.Context, cfg Config) (shutdown func(context.Context) error) {
	serviceName, version, deploymentEnv := cfg.ServiceName, cfg.Version, cfg.DeploymentEnv
	traceEndpoint, logEndpoint, metricEndpoint := cfg.TraceEndpoint, cfg.LogEndpoint, cfg.MetricEndpoint

	if traceEndpoint == "" && logEndpoint == "" && metricEndpoint == "" {
		return func(context.Context) error { return nil }
	}

	// Use a separate context with timeout so exporter creation doesn't block startup.
	setupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Datadog unified service tagging relies on service/env/version. Without
	// env and version everything lands in the default env and version-based
	// comparisons are unavailable.
	attrs := []attribute.KeyValue{semconv.ServiceName(serviceName)}
	if version != "" {
		attrs = append(attrs, semconv.ServiceVersion(version))
	}
	if deploymentEnv != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(deploymentEnv))
	}

	res, err := resource.New(setupCtx,
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		slog.Error("failed to create OTel resource", "error", err)
		return func(context.Context) error { return nil }
	}

	var shutdownFuncs []func(context.Context) error

	// Tracing
	if traceEndpoint != "" {
		traceExporter, err := otlptracehttp.New(setupCtx,
			otlptracehttp.WithEndpointURL(ensurePath(traceEndpoint, "/v1/traces")),
		)
		if err != nil {
			slog.Error("failed to create OTLP trace exporter", "error", err)
		} else {
			tp := sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(traceExporter),
				sdktrace.WithResource(res),
				sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplerRatio))),
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

	// Metrics
	if metricEndpoint != "" {
		metricExporter, err := otlpmetrichttp.New(setupCtx,
			otlpmetrichttp.WithEndpointURL(ensurePath(metricEndpoint, "/v1/metrics")),
		)
		if err != nil {
			slog.Error("failed to create OTLP metric exporter", "error", err)
		} else {
			mp := sdkmetric.NewMeterProvider(
				sdkmetric.WithResource(res),
				sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
			)
			otel.SetMeterProvider(mp)
			// Bind business-outcome instruments to this provider (not the
			// default no-op that was active at package init).
			metrics.Register()
			shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
			slog.Info("OpenTelemetry metrics enabled", "endpoint", metricEndpoint, "service", serviceName)
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
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
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
