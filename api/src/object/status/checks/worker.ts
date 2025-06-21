import type { Checker, CheckerInput, CheckerOutput } from '../registery.ts';

export default class implements Checker {
    async check(input: CheckerInput, result: CheckerOutput): Promise<void> {
        if (input.config.adminWorker.enableAdminWorker) {
            result.warning(
                'admin.worker',
                'Admin-Worker should be disabled',
                'Shopmon',
                'https://developer.shopware.com/docs/guides/plugins/plugins/framework/message-queue/add-message-handler#the-admin-worker',
            );
            return;
        }

        result.success('admin.worker', 'Admin-Worker is disabled');
    }
}
