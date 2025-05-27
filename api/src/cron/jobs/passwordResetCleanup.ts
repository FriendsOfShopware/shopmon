import { getConnection, schema } from '../../db.ts';
import { lt } from 'drizzle-orm';

export async function passwordResetCleanupJob() {
    const drizzle = getConnection();
    
    try {
        const now = new Date();
        
        // Count expired tokens before deleting
        const expiredTokens = await drizzle
            .select({ count: schema.passwordResetTokens.id })
            .from(schema.passwordResetTokens)
            .where(lt(schema.passwordResetTokens.expires, now));
        
        const count = expiredTokens.length;
        
        // Delete expired password reset tokens
        await drizzle
            .delete(schema.passwordResetTokens)
            .where(lt(schema.passwordResetTokens.expires, now))
            .execute();
        
        if (count > 0) {
            console.log(`Cleaned up ${count} expired password reset tokens`);
        }
    } catch (error) {
        console.error('Error cleaning up password reset tokens:', error);
        throw error;
    }
}