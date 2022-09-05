declare global {
    interface Env {
        DATABASE_HOST: string;
        DATABASE_USER: string;
        DATABASE_PASSWORD: string;
        MAIL_URL: string;
        MAIL_SECRET: string;
        MAIL_FROM: string;
        FRONTEND_URL: string;
        SENTRY_DSN: string;
        kvStorage: KVNamespace;
        SHOPS_SCRAPE: DurableObjectNamespace;
    }

    interface Team {
        id: string;
        ownerId: string;
    }

    interface Request {
        userId: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        params: any;
        team: Team;
    }
}

export {}