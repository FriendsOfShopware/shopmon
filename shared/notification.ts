export interface WebsocketMessage {
    notification?: Notification;
    shopUpdate?: ShopUpdate;
}

export interface ShopUpdate {
    id: number;
    team_id: number;
}

export interface NotificationCreation {
    level: 'error'|'warning';
    title: string;
    message: string;
    link: { name: string, params?: Record<string, string> }|false
}

export interface Notification extends NotificationCreation {
    id: number;
    read: boolean;
    created_at: string;
}