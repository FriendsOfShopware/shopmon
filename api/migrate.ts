import { migrate } from 'drizzle-orm/libsql/migrator';
import { createClient } from '@libsql/client';
import {drizzle} from 'drizzle-orm/libsql';
import { schema } from './src/db';

const connection = createClient({
  url: process.env.LIBSQL_URL as string,
  authToken: process.env.LIBSQL_AUTH_TOKEN as string
});

const db = drizzle(connection, { schema })

// This will run migrations on the database, skipping the ones already applied
await migrate(db, { migrationsFolder: './drizzle' });
