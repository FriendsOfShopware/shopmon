import type { Config } from 'drizzle-kit';
export default {
    schema: './src/db.ts',
    out: './drizzle',
    dialect: 'postgresql',
    dbCredentials: {
        url: process.env.DATABASE_URL || 'postgres://shopmon:shopmon@localhost:5432/shopmon',
    },
} satisfies Config;
