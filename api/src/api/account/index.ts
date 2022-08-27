import { Router } from "itty-router";
import { validateToken } from "../middleware/auth";
import { accountDelete, accountMe, accountUpdate, listUserShops } from "./me";

const accountRouter = Router({ base: "/api/account" });

accountRouter.get('/me', validateToken, accountMe);
accountRouter.get('/me/shops', validateToken, listUserShops)
accountRouter.patch('/me', validateToken, accountUpdate);
accountRouter.delete('/me', validateToken, accountDelete);

export default accountRouter;