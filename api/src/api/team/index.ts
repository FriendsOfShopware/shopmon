import {Router} from "itty-router";
import {validateToken} from "../middleware/auth";
import {validateTeam, validateTeamOwner} from "../middleware/team";
import {createShop} from "./create_shop";
import {updateShop} from "./update_shop";
import {deleteShop} from "./delete_shop";
import {getShop} from "./get_stop";
import {listShops} from "./list_shops";
import {addMember, listMembers, removeMember} from "./members";
import {refreshShop} from "./refresh_shop";
import {createTeam} from "./create_team";
import {updateTeam} from "./update_team";
import {deleteTeam} from "./delete_team";
import {clearShopCache} from "./clear_shop_cache";

const teamRouter = Router({ base: "/api/team" });

teamRouter.post('', validateToken, createTeam);
teamRouter.patch('/:teamId', validateToken, validateTeam, validateTeamOwner, updateTeam);
teamRouter.delete('/:teamId', validateToken, validateTeam, validateTeamOwner, deleteTeam);
teamRouter.get('/:teamId/members', validateToken, validateTeam, listMembers);
teamRouter.post('/:teamId/members', validateToken, validateTeam, validateTeamOwner, addMember);
teamRouter.delete('/:teamId/members/:userId', validateToken, validateTeam, validateTeamOwner, removeMember);
teamRouter.get('/:teamId/shops', validateToken, validateTeam, listShops);
teamRouter.post('/:teamId/shops', validateToken, validateTeam, validateTeamOwner, createShop);
teamRouter.get('/:teamId/shop/:shopId', validateToken, validateTeam, getShop);
teamRouter.patch('/:teamId/shop/:shopId', validateToken, validateTeam, updateShop);
teamRouter.post('/:teamId/shop/:shopId/refresh', validateToken, validateTeam, refreshShop);
teamRouter.delete('/:teamId/shop/:shopId', validateToken, validateTeam, deleteShop);
teamRouter.post('/:teamId/shop/:shopId/clear_cache', validateToken, validateTeam, clearShopCache);

export default teamRouter;