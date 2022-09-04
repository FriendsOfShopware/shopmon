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
}

interface Team {
    id: string;
    ownerId: string;
}

interface Request {
    userId: string;
    params: any;
    team: Team;
}