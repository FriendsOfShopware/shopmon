export interface User {
    id: number;
    username: string;
    email: string;
    created_at: string;
    avatar: string;

    teams: Team[];
}

export interface Team {
    id: number;
    name: string;
    created_at: string;
}