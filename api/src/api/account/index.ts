import { Router } from "itty-router";
import { validateToken } from "../middleware/auth";
import { deleteAllNotifications, deleteNotification, getNotifications, markAllReadNotification } from "./notification";

const accountRouter = Router({ base: "/api/account" });

accountRouter.get('/me/notifications', validateToken, getNotifications);
accountRouter.patch('/me/notifications/mark-all-read', validateToken, markAllReadNotification);
accountRouter.delete('/me/notifications/:id', validateToken, deleteNotification);
accountRouter.delete('/me/notifications', validateToken, deleteAllNotifications);

export default accountRouter;
