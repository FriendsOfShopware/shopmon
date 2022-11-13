import { getLatestShopwareVersion } from "./shopware_version";
import { checkExtensionCompatibility } from "./shopware_check_extension_compatibility"
import { Router } from "itty-router";

const infoRouter = Router({ base: "/api/info" });

infoRouter.get('/latest-shopware-version', getLatestShopwareVersion);
infoRouter.post('/check-extension-compatibility', checkExtensionCompatibility);

export default infoRouter