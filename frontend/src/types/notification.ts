export type WebsocketMessage = {
    notification?: Notification;
    shopUpdate?: ShopUpdate;
}

export type ShopUpdate = {
    id: number;
    team_id: number;
}

export type Notification = {
    id: number;
    read: boolean;
    created_at: string;
    level: 'error'|'warning';
    title: string;
    message: string;
    link: { name: string, params?: Record<string, string> }|false
}

export type NotificationCreation = Omit<Notification, 'id' | 'read' | 'created_at'>
