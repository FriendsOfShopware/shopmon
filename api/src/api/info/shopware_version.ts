import { JsonResponse } from "../common/response"

export async function getLatestShopwareVersion(): Promise<Response> {
    const installApiResp = await fetch('https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/all-supported-php-versions-by-shopware-version.json', {
        cf: {
            cacheEverything: true,
            cacheTtl: 60 * 60 * 2, // 2 hours
        },
        method: 'GET',
        headers: {
            Accept: 'application/json',
            'User-Agent': 'Shopmon'
        }
    });

    const installApiData = await installApiResp.json() as ShopwareVersion[]

    return new JsonResponse(installApiData);
}

interface ShopwareVersion {
    [version: string]: string[];
}
