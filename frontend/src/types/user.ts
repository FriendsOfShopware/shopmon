import type { Team } from './team'

export interface User {
    id: number;
    username: string;
    email: string;
    created_at: string;
    avatar: string;

    teams: Team[];
}
