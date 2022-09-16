import {cast, connect, Connection, Field} from '@planetscale/database/dist'

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
    return connect({
        cast: inflate,
        host: env.DATABASE_HOST,
        username: env.DATABASE_USER,
        password: env.DATABASE_PASSWORD,
    })
}