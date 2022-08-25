import { Router } from "itty-router";
import { validateToken } from "../middleware/auth";
import { accountDelete, accountMe } from "./me";

const accountRouter = Router({ base: "/api/account" });

accountRouter.get('/me', validateToken, accountMe);
accountRouter.delete('/me', validateToken, accountDelete);

export default accountRouter;