import { connect, Connection } from '@planetscale/database'

let workerConn: Connection|null = null;

const config = {
  host: globalThis.DATABASE_HOST,
  username: globalThis.DATABASE_USER,
  password: globalThis.DATABASE_PASSWORD,
}

export function getConnection(): Connection {
    if (workerConn !== null) {
        return workerConn
    }

    return connect(config)
}


export function getKv(): KVNamespace {
  return globalThis.kvStorage;
}