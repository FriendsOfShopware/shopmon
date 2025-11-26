import { logger } from '@sentry/node';
import { cleanupExpiredLocks } from '#src/modules/lock/lock.repository.ts';

/**
 * Cron job to clean up expired locks
 * This job should be scheduled to run periodically (e.g., daily)
 */
export async function lockCleanupJob() {
    try {
        const removedCount = await cleanupExpiredLocks();
        logger.info(
            `Lock cleanup job completed: ${removedCount} expired locks removed`,
        );
    } catch (error) {
        logger.error('Error in lock cleanup job', error);
    }
}
