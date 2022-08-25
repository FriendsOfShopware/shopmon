import { Router } from "itty-router";
import { validateToken } from "../middleware/auth";
import { validateTeam, validateTeamOwner } from "../middleware/team";
import { createShop } from "./create_shop";
import { deleteShop } from "./delete_shop";
import { getShop } from "./get_stop";
import { listShops } from "./list_shops";
import { addMember, listMembers, removeMember } from "./members";
import { deleteTeam } from "./team";

const teamRouter = Router({base: "/api/team"});

teamRouter.delete('/:teamId', validateToken, validateTeam, validateTeamOwner, deleteTeam);
teamRouter.get('/:teamId/members', validateToken, validateTeam, listMembers);
teamRouter.post('/:teamId/members', validateToken, validateTeam, validateTeamOwner, addMember);
teamRouter.delete('/:teamId/members/:userId', validateToken, validateTeam, validateTeamOwner, removeMember);
teamRouter.get('/:teamId/shops', validateToken, validateTeam, listShops);
teamRouter.post('/:teamId/shops', validateToken, validateTeam, validateTeamOwner, createShop);
teamRouter.get('/:teamId/shop/:shopId', validateToken, validateTeam, getShop);
teamRouter.delete('/:teamId/shop/:shopId', validateToken, validateTeam, deleteShop);

export default teamRouter;