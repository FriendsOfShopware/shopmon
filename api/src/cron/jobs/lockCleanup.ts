import { logger } from '@sentry/bun';
import LockRepository from '../../repository/lock.js';

/**
 * Cron job to clean up expired locks
 * This job should be scheduled to run periodically (e.g., daily)
 */
export async function lockCleanupJob() {
    try {
        const removedCount = await LockRepository.cleanupExpiredLocks();
        logger.info(`Lock cleanup job completed: ${removedCount} expired locks removed`);
    } catch (error) {
        logger.error('Error in lock cleanup job', error);
    }
}
