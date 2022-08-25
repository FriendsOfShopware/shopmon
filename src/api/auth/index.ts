import { Router } from "itty-router";
import { login } from "./login";
import register from "./register";

const authRouter = Router({base: "/api/auth"});

authRouter.post("/register", register);
authRouter.post("/login", login);

export default authRouter