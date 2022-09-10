import type { User } from './user';

export interface WebsocketMessage {
    notification?: Notification;
    shopUpdate?: ShopUpdate;
}

export interface ShopUpdate {
    id: number;
    team_id: number;
}

export interface Notification {
    id?: number;
    level: 'error'|'warning';
    title: string;
    message: string;
    link: { name: string, params?: Record<string, string> }|false
    read?: boolean;
}