import { lt } from 'drizzle-orm';
import { getConnection, schema } from '../../db.js';

export async function sessionCleanupJob() {
    const con = getConnection();

    // First count the expired sessions
    const expiredSessions = await con
        .select({ count: schema.sessions.id })
        .from(schema.sessions)
        .where(lt(schema.sessions.expires, new Date()));

    const count = expiredSessions.length;

    // Delete all sessions that have expired
    await con
        .delete(schema.sessions)
        .where(lt(schema.sessions.expires, new Date()))
        .execute();

    console.log(`Cleaned up ${count} expired sessions`);
}
