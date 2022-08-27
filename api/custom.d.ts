declare var DATABASE_HOST: string;
declare var DATABASE_USER: string;
declare var DATABASE_PASSWORD: string;
declare var MAIL_URL: string;
declare var MAIL_SECRET: string;
declare var MAIL_FROM: string;
declare var FRONTEND_URL: string;
declare var kvStorage: KVNamespace;

interface Team {
    id: string;
    ownerId: string;
}

interface Request {
    userId: string;
    params: any;
    team: Team;
}