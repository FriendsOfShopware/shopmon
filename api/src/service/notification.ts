import type { Drizzle } from '#src/db.ts';
import NotificationRepository from '#src/repository/notifications.ts';

export const listNotifications = async (db: Drizzle, userId: string) => {
    return await NotificationRepository.findAllByUserId(db, userId);
};

export const deleteNotification = async (
    db: Drizzle,
    userId: string,
    notificationId?: number,
) => {
    if (notificationId) {
        await NotificationRepository.deleteById(db, userId, notificationId);
        return true;
    }

    await NotificationRepository.deleteAllByUserId(db, userId);
    return true;
};

export const markAllRead = async (db: Drizzle, userId: string) => {
    await NotificationRepository.markAllRead(db, userId);
};
