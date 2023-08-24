declare global {
    const SENTRY_RELEASE: string;

    interface Env {
        DATABASE_HOST: string;
        DATABASE_USER: string;
        DATABASE_PASSWORD: string;
        MAILGUN_KEY: string;
        MAILGUN_DOMAIN: string;
        MAIL_FROM: string;
        FRONTEND_URL: string;
        SENTRY_DSN: string;
        PAGESPEED_API_KEY: string;
        APP_SECRET: string;
        DISABLE_REGISTRATION: boolean;
        kvStorage: KVNamespace;
        SHOPS_SCRAPE: DurableObjectNamespace;
        PAGESPEED_SCRAPE: DurableObjectNamespace;
        USER_SOCKET: DurableObjectNamespace;
        FILES: R2Bucket;
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