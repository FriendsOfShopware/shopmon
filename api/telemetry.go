package main

import (
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/telemetry"
)

// telemetryConfig translates the OTel environment configuration into the
// telemetry setup config. serviceName is the process-specific APM name:
// the HTTP server uses ${OTEL_SERVICE_NAME}-api and the worker uses
// ${OTEL_SERVICE_NAME}-worker so both can share one base env value.
func telemetryConfig(cfg *config.Config, serviceName string) telemetry.Config {
	return telemetry.Config{
		ServiceName:    serviceName,
		Version:        cfg.OtelServiceVersion,
		DeploymentEnv:  cfg.OtelDeploymentEnv,
		TraceEndpoint:  cfg.OtelTraceEndpoint,
		LogEndpoint:    cfg.OtelLogEndpoint,
		MetricEndpoint: cfg.OtelMetricEndpoint,
		SamplerRatio:   cfg.OtelSamplerRatio,
	}
}
