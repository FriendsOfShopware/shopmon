import register from "./api/auth/register";
import { login } from "./api/auth/login";
import { accountMe } from "./api/account/me";
import { Router } from "itty-router";
import { validateToken } from "./api/middleware/auth";
import { validateTeam } from "./api/middleware/team";
import { listShops } from "./api/team/list_shops";
import { createShop } from "./api/team/create_shop";
import { getShop } from "./api/team/get_stop";
import { deleteShop } from "./api/team/delete_shop";

const router = Router();

router.post("/api/auth/register", register);
router.post("/api/auth/login", login);
router.get("/api/account/me", validateToken, accountMe);
router.get('/api/teams/:teamId/shops', validateToken, validateTeam, listShops);
router.post('/api/teams/:teamId/shops', validateToken, validateTeam, createShop);
router.get('/api/teams/:teamId/shop/:shopId', validateToken, validateTeam, getShop);
router.delete('/api/teams/:teamId/shop/:shopId', validateToken, validateTeam, deleteShop);

router.all('*', () => new Response('Not Found.', { status: 404 }));

export default router;