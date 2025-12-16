import { migrate } from 'drizzle-orm/postgres-js/migrator';
import { closeConnection, getConnection } from '#src/db.ts';

async function runMigrations() {
    console.log('Running migrations...');

    // This will run migrations on the database, skipping the ones already applied
    await migrate(getConnection(), { migrationsFolder: './drizzle' });

    console.log('Migrations completed successfully!');
    await closeConnection();
}

runMigrations();
