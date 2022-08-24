declare var DATABASE_HOST: string;
declare var DATABASE_USER: string;
declare var DATABASE_PASSWORD: string;
declare var kvStorage: KVNamespace;

interface Request {
    userId: string;
    params: any;
}