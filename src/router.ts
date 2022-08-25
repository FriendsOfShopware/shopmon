import { Router } from "itty-router";
import teamRouter from './api/team';
import authRouter from "./api/auth";
import accountRouter from "./api/account";
import { JsonResponse } from "./api/common/response";

const router = Router();

router.all("/api/account/*", accountRouter.handle);
router.all('/api/team/*', teamRouter.handle);
router.all('/api/auth/*', authRouter.handle);

router.options('*', () => new Response("", {headers: {
    "Access-Control-Allow-Origin": "https://app.swaggerhub.com",
    "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
    "Access-Control-Allow-Credentials": "true",
    "Access-Control-Allow-Headers": "Content-Type, token"}
}));
router.all('*', () => new JsonResponse({ message: 'Not found' }, 404));

export default router;