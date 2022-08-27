import { Router } from "itty-router";
import { confirmMail } from "./confirm";
import { login } from "./login";
import register from "./register";
import { confirmResetPassword, resetAvailable, resetPasswordMail } from "./reset";

const authRouter = Router({base: "/api/auth"});

authRouter.post("/register", register);
authRouter.post("/login", login);
authRouter.post('/confirm/:token', confirmMail);
authRouter.post('/reset', resetPasswordMail)
authRouter.get('/reset/:token', resetAvailable)
authRouter.post('/reset/:token', confirmResetPassword);

export default authRouter