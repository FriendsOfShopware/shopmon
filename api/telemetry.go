package main

import (
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/telemetry"
)

// telemetryConfig translates the OTel environment configuration into the
// telemetry setup config. The service name is passed in because the worker
// reports under its own name.
func telemetryConfig(cfg *config.Config, serviceName string) telemetry.Config {
	return telemetry.Config{
		ServiceName:   serviceName,
		Version:       cfg.OtelServiceVersion,
		DeploymentEnv: cfg.OtelDeploymentEnv,
		TraceEndpoint: cfg.OtelTraceEndpoint,
		LogEndpoint:   cfg.OtelLogEndpoint,
		SamplerRatio:  cfg.OtelSamplerRatio,
	}
}
