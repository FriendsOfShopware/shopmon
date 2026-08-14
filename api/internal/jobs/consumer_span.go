package jobs

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// annotateConsumerEnvironment copies the environment id onto the active
// go-queue Consumer span (EnvironmentScrape process / SitespeedScrape process).
// Child Internal spans already carry richer environment attributes; Datadog
// facets on the entry span need this low-cardinality id without opening
// children. Do not add high-cardinality fields (URL, emails, extension lists).
func annotateConsumerEnvironment(ctx context.Context, environmentID int32) {
	trace.SpanFromContext(ctx).SetAttributes(
		attribute.Int("environment.id", int(environmentID)),
	)
}

// annotateConsumerOutcome stamps a tiny ok/error enum on the active Consumer
// span after the handler returns. Mirrors go-queue otel status (error vs ok)
// as a facetable attribute; business-specific outcomes stay on metrics/child spans.
func annotateConsumerOutcome(ctx context.Context, err error) {
	outcome := "ok"
	if err != nil {
		outcome = "error"
	}
	trace.SpanFromContext(ctx).SetAttributes(
		attribute.String("outcome", outcome),
	)
}

// runEnvironmentJob annotates the Consumer span with environment.id up front
// and outcome when the job finishes, then returns the job error unchanged.
// Do not RecordError / SetStatus here: queueotel.Middleware already does that
// on the Consumer span after the handler returns; duplicating it would emit
// two exception events for the same failure.
func runEnvironmentJob(ctx context.Context, environmentID int32, run func(context.Context) error) error {
	annotateConsumerEnvironment(ctx, environmentID)
	err := run(ctx)
	annotateConsumerOutcome(ctx, err)
	return err
}
