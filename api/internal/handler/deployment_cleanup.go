package handler

import (
	"context"
	"time"
)

const deploymentOutputCleanupTimeout = 30 * time.Second

func deploymentOutputCleanupContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), deploymentOutputCleanupTimeout)
}
