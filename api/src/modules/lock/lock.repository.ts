import { eq, lt } from 'drizzle-orm';
import { getConnection, lock } from '#src/db.ts';

export async function isLocked(key: string): Promise<boolean> {
    const con = getConnection();
    const result = await con.select().from(lock).where(eq(lock.key, key)).get();

    if (!result) {
        return false;
    }

    if (result.expires < new Date()) {
        await con.delete(lock).where(eq(lock.key, key)).execute();
        return false;
    }

    return true;
}

export async function createLock(key: string, ttl = 3600): Promise<void> {
    const con = getConnection();
    const expires = new Date(Date.now() + ttl * 1000);

    await con
        .insert(lock)
        .values({
            key,
            expires,
            createdAt: new Date(),
        })
        .onConflictDoUpdate({
            target: lock.key,
            set: {
                expires,
                createdAt: new Date(),
            },
        })
        .execute();
}

export async function cleanupExpiredLocks(): Promise<void> {
    const con = getConnection();
    await con.delete(lock).where(lt(lock.expires, new Date())).execute();
}
