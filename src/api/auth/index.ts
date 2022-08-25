import { Router } from "itty-router";
import { confirmMail } from "./confirm";
import { login } from "./login";
import register from "./register";

const authRouter = Router({base: "/api/auth"});

authRouter.post("/register", register);
authRouter.post("/login", login);
authRouter.post('/confirm/:token', confirmMail);

export default authRouter