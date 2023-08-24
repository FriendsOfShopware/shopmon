import { cast, connect, Connection, Field } from '@planetscale/database'

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
        // eslint-disable-next-line @typescript-eslint/no-explicit-any -- remove when fixed https://github.com/cloudflare/workerd/issues/698
        fetch: async (url: string, init: any) => {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
            delete init.cache
            // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
            return await fetch(url, init)
        },
    })
}