import { and, eq, gt, lt } from 'drizzle-orm';
import { getConnection, lock } from '../db.js';

/**
 * Repository for managing locks
 * Used to ensure certain operations are only executed once within a specified time period
 */
export default class LockRepository {
    /**
     * Check if a lock exists and is not expired
     * @param key The unique key to check
     * @returns True if the lock exists and is not expired, false otherwise
     */
    static async isLocked(key: string): Promise<boolean> {
        const con = getConnection();
        const now = new Date();
        const result = await con
            .select()
            .from(lock)
            .where(
                and(
                    eq(lock.key, key),
                    // Only consider locks that haven't expired
                    gt(lock.expires, now)
                )
            )
            .get();

        return !!result;
    }

    /**
     * Set a lock with a specified expiration time
     * @param key The unique key for the lock
     * @param expiresInSeconds Number of seconds until the lock expires (default: 24 hours)
     * @returns True if the lock was successfully set, false if it already exists and is not expired
     */
    static async createLock(key: string, expiresInSeconds: number = 86400): Promise<boolean> {
        // First check if the lock already exists and is not expired
        if (await this.isLocked(key)) {
            return false;
        }

        const con = getConnection();
        const now = new Date();
        const expires = new Date(now.getTime() + expiresInSeconds * 1000);

        // Insert or replace the lock
        await con
            .insert(lock)
            .values({
                key,
                expires,
                createdAt: now,
            })
            .onConflictDoUpdate({
                target: lock.key,
                set: {
                    expires,
                    createdAt: now,
                },
            });

        return true;
    }

    /**
     * Remove a specific lock
     * @param key The unique key to remove
     */
    static async removeLock(key: string): Promise<void> {
        const con = getConnection();
        await con.delete(lock).where(eq(lock.key, key));
    }

    /**
     * Clean up all expired locks
     * @returns The number of locks that were removed
     */
    static async cleanupExpiredLocks(): Promise<number> {
        const con = getConnection();
        const now = new Date();
        
        // Delete all locks that have expired
        const result = await con
            .delete(lock)
            .where(lt(lock.expires, now))
            .returning();

        return result.length;
    }
}
