declare var DATABASE_HOST: string;
declare var DATABASE_USER: string;
declare var DATABASE_PASSWORD: string;
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