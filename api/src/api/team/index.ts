import { Router } from "itty-router";
import { validateToken } from "../middleware/auth";
import { validateTeam, validateTeamOwner } from "../middleware/team";
import { createShop } from "./create_shop";
import { updateShop } from "./update_shop";
import { deleteShop } from "./delete_shop";
import { getShop } from "./get_shop";
import { listShops } from "./list_shops";
import { refreshShop } from "./refresh_shop";
import { clearShopCache } from "./clear_shop_cache";
import { reScheduleTask } from "./reschedule_task";
import { validateShop } from "../middleware/shop";
import { shopImage } from "./shop_image";

const teamRouter = Router({ base: "/api/team" });

teamRouter.get('/:teamId/shops', validateToken, validateTeam, listShops);
teamRouter.post('/:teamId/shops', validateToken, validateTeam, validateTeamOwner, createShop);
teamRouter.get('/:teamId/shop/:shopId', validateToken, validateTeam, validateShop, getShop);
teamRouter.patch('/:teamId/shop/:shopId', validateToken, validateTeam, updateShop, validateShop);
teamRouter.post('/:teamId/shop/:shopId/refresh', validateToken, validateTeam, validateShop, refreshShop);
teamRouter.delete('/:teamId/shop/:shopId', validateToken, validateTeam, validateShop, validateTeamOwner, deleteShop);
teamRouter.post('/:teamId/shop/:shopId/clear_cache', validateToken, validateTeam, validateShop, clearShopCache);
teamRouter.post('/:teamId/shop/:shopId/reschedule_task/:taskId', validateToken, validateTeam, validateShop, reScheduleTask);
teamRouter.get('/pagespeed/:uuid/screenshot.jpg', shopImage);

export default teamRouter;
