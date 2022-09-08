import type { User } from './user';

export interface Notification {
    user_id: User['id'];
    level: 'error'|'warning';
    title: string;
    message: string;
    link: {
        name: string;
        params: Record<string, string>;
    }
}