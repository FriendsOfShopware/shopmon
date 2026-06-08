package checker

import "context"

func checkWorker(_ context.Context, input Input, output *Output) {
	if input.Config.AdminWorker.EnableAdminWorker {
		output.Warning("admin.worker",
			"The admin worker is enabled. This can cause performance issues. Consider using the CLI worker instead.",
			"Shopware",
			"https://developer.shopware.com/docs/guides/hosting/infrastructure/message-queue.html")
	} else {
		output.Success("admin.worker",
			"Admin worker is disabled. CLI worker is being used.",
			"Shopware", "")
	}
}
