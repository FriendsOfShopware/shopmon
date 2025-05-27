import type { Config } from 'drizzle-kit';
export default {
    schema: './src/db.ts',
    out: './drizzle',
    dialect: 'sqlite',
} satisfies Config;
