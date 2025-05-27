import { migrate } from 'drizzle-orm/bun-sqlite/migrator';
import { closeConnection, getConnection } from './src/db.ts';

async function runMigrations() {
    console.log('Running migrations...');

    try {
        // This will run migrations on the database, skipping the ones already applied
        await migrate(getConnection(), { migrationsFolder: './drizzle' });

        console.log('Migrations completed successfully!');
        process.exit(0);
    } catch (error) {
        console.error('Migration failed:', error);
        process.exit(1);
    } finally {
        // Close the database connection properly
        closeConnection();
    }
}

runMigrations();
