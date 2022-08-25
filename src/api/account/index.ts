import { Router } from "itty-router";
import { validateToken } from "../middleware/auth";
import { accountMe } from "./me";

const accountRouter = Router({ base: "/api/account" });

accountRouter.get("/me", validateToken, accountMe);

export default accountRouter;