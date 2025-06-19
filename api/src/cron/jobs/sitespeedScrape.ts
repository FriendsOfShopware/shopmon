import { eq } from 'drizzle-orm';
import { getConnection, schema } from '../../db.ts';
import { sanitizeSitespeedLabel } from '../../util.ts';

export async function scrapeSitespeedForAllShops() {
    const drizzle = getConnection();

    console.log('Starting sitespeed scrape for all shops');

    const shops = await drizzle.query.shop.findMany({
        columns: {
            id: true,
        },
        where: eq(schema.shop.sitespeedEnabled, true),
    });

    console.log(`Found ${shops.length} shops with sitespeed enabled`);

    for (const shop of shops) {
        await scrapeSingleSitespeedShop(shop.id);
    }

    console.log('Sitespeed scrape completed for all shops');
}

export async function scrapeSingleSitespeedShop(shopId: number) {
    const drizzle = getConnection();

    const shop = await drizzle.query.shop.findFirst({
        columns: {
            id: true,
            url: true,
            name: true,
            sitespeedEnabled: true,
            sitespeedUrls: true,
        },
        where: eq(schema.shop.id, shopId),
    });

    if (!shop) {
        console.error(`Shop ${shopId} not found`);
        return;
    }

    if (!shop.sitespeedEnabled) {
        console.log(`Sitespeed is disabled for shop ${shop.id}: ${shop.name}`);
        return;
    }

    console.log(`Running sitespeed analysis for shop ${shop.id}: ${shop.name}`);

    const sitespeedServiceUrl = process.env.APP_SITESPEED_ENDPOINT;

    const urlsToTest =
        shop.sitespeedUrls.length > 0
            ? shop.sitespeedUrls.slice(0, 5)
            : [{ url: shop.url, label: 'Homepage' }];

    try {
        const response = await fetch(`${sitespeedServiceUrl}/analyze`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                shopId: shop.id,
                urls: urlsToTest.map((urlConfig) => urlConfig.url),
            }),
        });

        const result = (await response.json()) as {
            metrics: {
                ttfb?: number;
                fullyLoaded?: number;
                largestContentfulPaint?: number;
                firstContentfulPaint?: number;
                cumulativeLayoutShift?: number;
                speedIndex?: number;
                transferSize?: number;
            };
        };

        // Store the metrics in the database
        await drizzle
            .insert(schema.shopSitespeed)
            .values({
                shopId: shop.id,
                url: urlsToTest[0].url,
                label: urlsToTest[0].label,
                createdAt: new Date(),
                ttfb: result.metrics.ttfb || null,
                fullyLoaded: result.metrics.fullyLoaded || null,
                largestContentfulPaint:
                    result.metrics.largestContentfulPaint || null,
                firstContentfulPaint:
                    result.metrics.firstContentfulPaint || null,
                cumulativeLayoutShift:
                    result.metrics.cumulativeLayoutShift || null,
                speedIndex: result.metrics.speedIndex || null,
                transferSize: result.metrics.transferSize || null,
            })
            .execute();

        console.log(`Sitespeed analysis completed for shop ${shop.id}`);
    } catch (error) {
        console.error(
            `Error running sitespeed analysis for shop ${shop.id}`,
            error,
        );
        throw error;
    }
}
