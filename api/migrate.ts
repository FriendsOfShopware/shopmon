import { migrate } from 'drizzle-orm/libsql/migrator';
import { getConnection } from './src/db.ts';

async function runMigrations() {
    console.log('Running migrations...');

    // This will run migrations on the database, skipping the ones already applied
    await migrate(getConnection(false), { migrationsFolder: './drizzle' });

    console.log('Migrations completed successfully!');
    process.exit(0);
}

runMigrations();
