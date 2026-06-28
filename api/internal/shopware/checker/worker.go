package checker

import "context"

func checkWorker(_ context.Context, input Input, output *Output) {
	if input.Config.AdminWorker.EnableAdminWorker {
		output.Warning("admin.worker", "check.worker.enabled", nil, "Shopware",
			"https://developer.shopware.com/docs/guides/hosting/infrastructure/message-queue.html")
	} else {
		output.Success("admin.worker", "check.worker.disabled", nil, "Shopware", "")
	}
}
