import { OTLPLogExporter } from '@opentelemetry/exporter-logs-otlp-proto/build/src/platform/node/index.js';
import { OTLPMetricExporter } from '@opentelemetry/exporter-metrics-otlp-proto';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-proto';
import { PinoInstrumentation } from '@opentelemetry/instrumentation-pino';
import { UndiciInstrumentation } from '@opentelemetry/instrumentation-undici';
import { SimpleLogRecordProcessor } from '@opentelemetry/sdk-logs';
import { PeriodicExportingMetricReader } from '@opentelemetry/sdk-metrics';
import { NodeSDK } from '@opentelemetry/sdk-node';
import { AlwaysOnSampler } from '@opentelemetry/sdk-trace-node';
import { BetterSqlite3Instrumentation } from 'opentelemetry-plugin-better-sqlite3';

const sdk = new NodeSDK({
    serviceName: 'shopmon',
    sampler: new AlwaysOnSampler(),
    traceExporter: new OTLPTraceExporter(),
    metricReader: new PeriodicExportingMetricReader({
        exporter: new OTLPMetricExporter(),
    }),
    logRecordProcessors: [new SimpleLogRecordProcessor(new OTLPLogExporter())],
    instrumentations: [
        new UndiciInstrumentation(),
        new PinoInstrumentation(),
        new BetterSqlite3Instrumentation(),
    ],
});

sdk.start();
