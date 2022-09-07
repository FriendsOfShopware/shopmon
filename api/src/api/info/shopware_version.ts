import { JsonResponse } from "../common/response"

export async function getLatestShopwareVersion(): Promise<Response> {
    const installApiResp = await fetch('https://update-api.shopware.com/v1/releases/install?major=6', {
        cf: {
            cacheEverything: true,
            cacheTtl: 60 * 60 * 2, // 2 hours
        }
    })

    const installApiData = await installApiResp.json() as ShopwareVersion[]
    
    return new JsonResponse(installApiData[0].version);
}

interface ShopwareVersion {
    version: string;
}