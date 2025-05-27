import { migrate } from 'drizzle-orm/bun-sqlite/migrator';
import { getConnection } from './src/db.ts';


console.log('Running migrations...');

// This will run migrations on the database, skipping the ones already applied
await migrate(getConnection(), { migrationsFolder: './drizzle' });

console.log('Migrations completed successfully!');
process.exit(0);
