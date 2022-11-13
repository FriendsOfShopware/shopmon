import { Extension } from "../../../../shared/shop";
import { JsonResponse } from "../common/response"
import { ErrorResponse } from "../common/response";

export async function checkExtensionCompatibility(req: Request): Promise<Response> {

    const { currentVersion, futureVersion, plugins } = await req.json() as { currentVersion: string, futureVersion: string, plugins: Extension[] };

    if ( !currentVersion ) {
        return new ErrorResponse('Current Version is empty', 400);
    }

    if ( !futureVersion ) {
        return new ErrorResponse('Future Version is empty', 400);
    }

    if ( plugins.length === 0 ) {
        return new ErrorResponse('Plugins is empty', 400);
    }

    const url = new URL('https://api.shopware.com/swplatform/autoupdate')
    url.searchParams.set('language', 'en-GB');
    url.searchParams.set('shopwareVersion', currentVersion);

    const checkExtensionCompatibilityApiResp = await fetch(url.toString(), {
        cf: {
            cacheEverything: true,
            cacheTtl: 60 * 60 * 2, // 2 hours
        },
        method: 'POST',
        body: JSON.stringify({
            "futureShopwareVersion": futureVersion,
            "plugins": plugins
        })
    })

    const checkExtensionCompatibilityApiData = await checkExtensionCompatibilityApiResp.json() as ShopwareExtensionCompatibility[]
    return new JsonResponse(checkExtensionCompatibilityApiData); 
}

interface ShopwareExtensionCompatibility {
    "name": string,
    "label": string,
    "iconPath": string,
    "status": {
        "name": string,
        "label": string,
        "type": string
    }
}