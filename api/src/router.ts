import { Router } from "itty-router";
import teamRouter from './api/team';
import authRouter from "./api/auth";
import accountRouter from "./api/account";
import { JsonResponse } from "./api/common/response";

const router = Router();

router.all("/api/account/*", accountRouter.handle);
router.all('/api/team/*', teamRouter.handle);
router.all('/api/auth/*', authRouter.handle);
router.all('*', () => new JsonResponse({ message: 'Not found' }, 404));

export default router;