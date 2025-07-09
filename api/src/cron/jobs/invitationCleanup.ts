import { eq, or } from 'drizzle-orm';
import { getConnection, invitation } from '../../db.ts';

export async function invitationCleanupJob() {
    const db = getConnection();

    await db
        .delete(invitation)
        .where(
            or(
                eq(invitation.status, 'canceled'),
                eq(invitation.status, 'accepted'),
            ),
        );
}
