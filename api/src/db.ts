import { connect, Connection } from '@planetscale/database'

let workerConn: Connection|null = null;

export function getConnection(env: Env): Connection {
  return connect({
    host: env.DATABASE_HOST,
    username: env.DATABASE_USER,
    password: env.DATABASE_PASSWORD,
  })
}