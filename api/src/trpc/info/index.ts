import { router, publicProcedure } from '..';
import { z } from 'zod';

export const infoRouter = router({
    checkExtensionCompatibility: publicProcedure
        .input(
            z.object({
                currentVersion: z.string(),
                futureVersion: z.string(),
                extensions: z.array(
                    z.object({
                        name: z.string(),
                        version: z.string(),
                    }),
                ),
            }),
        )
        .query(async ({ input }) => {
            const url = new URL(
                'https://api.shopware.com/swplatform/autoupdate',
            );
            url.searchParams.set('language', 'en-GB');
            url.searchParams.set('shopwareVersion', input.currentVersion);

            const checkExtensionCompatibilityApiResp = await fetch(
                url.toString(),
                {
                    cf: {
                        cacheEverything: true,
                        cacheTtl: 60 * 60 * 2, // 2 hours
                    },
                    method: 'POST',
                    body: JSON.stringify({
                        futureShopwareVersion: input.futureVersion,
                        plugins: input.extensions,
                    }),
                },
            );

            return (await checkExtensionCompatibilityApiResp.json()) as ShopwareExtensionCompatibility[];
        }),
    getLatestShopwareVersion: publicProcedure.query(async () => {
        const installApiResp = await fetch(
            'https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/all-supported-php-versions-by-shopware-version.json',
            {
                cf: {
                    cacheEverything: true,
                    cacheTtl: 60 * 60 * 2, // 2 hours
                },
                method: 'GET',
                headers: {
                    Accept: 'application/json',
                    'User-Agent': 'Shopmon',
                },
            },
        );

        return (await installApiResp.json()) as ShopwareVersion[];
    }),
});

interface ShopwareExtensionCompatibility {
    name: string;
    label: string;
    iconPath: string;
    status: {
        name: string;
        label: string;
        type: string;
    };
}

interface ShopwareVersion {
    [version: string]: string[];
}
