import { Router } from "itty-router";
import { confirmMail } from "./confirm";
import register from "./register";
import { confirmResetPassword, resetAvailable, resetPasswordMail } from "./reset";
import oauth from './oauth';

const authRouter = Router({ base: "/api/auth" });

authRouter.post('/token', oauth);
authRouter.post("/register", register);
authRouter.post('/confirm/:token', confirmMail);
authRouter.post('/reset', resetPasswordMail)
authRouter.get('/reset/:token', resetAvailable)
authRouter.post('/reset/:token', confirmResetPassword);

export default authRouter
