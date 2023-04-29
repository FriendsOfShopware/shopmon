import { Router } from "itty-router";
import { validateToken } from "../middleware/auth";
import { accountDelete, accountMe, accountUpdate, listUserShops, listUserChangelogs } from "./me";
import { deleteAllNotifications, deleteNotification, getNotifications, markAllReadNotification } from "./notification";

const accountRouter = Router({ base: "/api/account" });

accountRouter.get('/me', validateToken, accountMe);
accountRouter.get('/me/notifications', validateToken, getNotifications);
accountRouter.patch('/me/notifications/mark-all-read', validateToken, markAllReadNotification);
accountRouter.delete('/me/notifications/:id', validateToken, deleteNotification);
accountRouter.delete('/me/notifications', validateToken, deleteAllNotifications);
accountRouter.get('/me/shops', validateToken, listUserShops);
accountRouter.get('/me/changelogs', validateToken, listUserChangelogs);
accountRouter.patch('/me', validateToken, accountUpdate);
accountRouter.delete('/me', validateToken, accountDelete);

export default accountRouter;