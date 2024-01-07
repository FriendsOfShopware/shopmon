declare global {
    const SENTRY_RELEASE: string;

    interface Env {
        MAIL_ACTIVE: 'true' | 'false';
        MAIL_FROM: string;
        MAIL_FROM_NAME: string;

        MAIL_DKIM_DOMAIN: string;
        MAIL_DKIM_SELECTOR: string;
        MAIL_DKIM_PRIVATE_KEY: string | undefined;

        FRONTEND_URL: string;
        SENTRY_DSN: string;
        PAGESPEED_API_KEY: string;
        APP_SECRET: string;
        DISABLE_REGISTRATION: boolean;

        // cloudflare bindings
        kvStorage: KVNamespace;
        SHOPS_SCRAPE: DurableObjectNamespace;
        PAGESPEED_SCRAPE: DurableObjectNamespace;
        USER_SOCKET: DurableObjectNamespace;
        FILES: R2Bucket;
        shopmonDB: D1Database;
        sendMail: SendEmail;
    }

    interface Team {
        id: string;
        ownerId: number;
    }

    interface Request {
        userId: number;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        params: any;
        team: Team;
    }
}

export { }
