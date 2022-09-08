import { Router } from "itty-router";
import { validateToken } from "../middleware/auth";
import { accountDelete, accountMe, accountUpdate, listUserShops } from "./me";
import { deleteAllNotifications, getNotifications } from "./notification";

const accountRouter = Router({ base: "/api/account" });

accountRouter.get('/me', validateToken, accountMe);
accountRouter.get('/me/notifications', validateToken, getNotifications);
accountRouter.delete('/me/notifications', validateToken, deleteAllNotifications);
accountRouter.get('/me/shops', validateToken, listUserShops)
accountRouter.patch('/me', validateToken, accountUpdate);
accountRouter.delete('/me', validateToken, accountDelete);

export default accountRouter;