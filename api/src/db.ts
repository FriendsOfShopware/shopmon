import { cast, connect, Connection, Field } from '@planetscale/database/dist'
import { Database } from "@cloudflare/d1";

function inflate(field: Field, value: string|null) {
    if (field.type === 'DATETIME') {
        if (value === null) {
            return null
        }

        return new Date(value).toISOString()
    }
    return cast(field, value)
}


export function getConnection(env: Env): Connection {
    if (env.USE_LOCAL_DATABASE) {
        return new D1Wrapper(env.__D1_BETA__LOCAL_BINDING as never) as never;
    }

    return connect({
        cast: inflate,
        host: env.DATABASE_HOST,
        username: env.DATABASE_USER,
        password: env.DATABASE_PASSWORD,
    })
}

class D1Wrapper {

    public isLocalConnection = true;

    private db: Database;
    constructor(db: Database) {
        this.db = db;
    }

    public async execute(query: string, params: any[] = []) {
        try {
            const queryResult = await this.db.prepare(query).bind(...params).all() as any;
            return {
                rows: queryResult.results,
                insertId: queryResult.meta.last_row_id.toString(),
                ...queryResult
            }
        } catch (e) {
            console.error(e);
            throw e;
        }
    }
}