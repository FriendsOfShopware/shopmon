package main

import (
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
)

// busConfig translates the queue-related environment configuration into the
// job bus configuration shared by the server, worker, and fixture commands.
func busConfig(cfg *config.Config) jobs.BusConfig {
	return jobs.BusConfig{
		OTelEnabled: cfg.OtelEnabled,
		Driver:      cfg.QueueTransport,
		AMQP: jobs.AMQPConfig{
			DSN:             cfg.QueueAMQP.DSN,
			Exchange:        cfg.QueueAMQP.Exchange,
			Queue:           cfg.QueueAMQP.Queue,
			PrefetchCount:   cfg.QueueAMQP.PrefetchCount,
			DelayedExchange: cfg.QueueAMQP.DelayedExchange,
		},
	}
}
