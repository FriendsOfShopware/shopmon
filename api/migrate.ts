import { migrate } from 'drizzle-orm/libsql/migrator';
import { getConnection, schema } from './src/db.ts';


console.log('Running migrations...');

// This will run migrations on the database, skipping the ones already applied
await migrate(getConnection(), { migrationsFolder: './drizzle' });

console.log('Migrations completed successfully!');
process.exit(0);
