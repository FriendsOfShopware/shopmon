import { connect, Connection, cast, Field } from '@planetscale/database'

function inflate(field: Field, value: any) {
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