import { getLatestShopwareVersion } from "./shopware_version";
import { Router } from "itty-router";

const infoRouter = Router({ base: "/api/info" });

infoRouter.get('/latest-shopware-version', getLatestShopwareVersion);

export default infoRouter