import { eq } from 'drizzle-orm';
import { getConnection, schema } from '../../db.ts';

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

    if (!shop.sitespeedEnabled || !shop.sitespeedUrls) {
        console.log(
            `Sitespeed is disabled for shop ${shop.id}: ${shop.name} or no URLs configured`,
        );
        return;
    }

    console.log(`Running sitespeed analysis for shop ${shop.id}: ${shop.name}`);

    const sitespeedServiceUrl = process.env.APP_SITESPEED_ENDPOINT;
    try {
        const response = await fetch(`${sitespeedServiceUrl}/analyze`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                shopId: shop.id,
                urls: shop.sitespeedUrls,
            }),
        });

        const result = (await response.json()) as {
            ttfb: number;
            fullyLoaded: number;
            largestContentfulPaint: number;
            firstContentfulPaint: number;
            cumulativeLayoutShift: number;
            transferSize: number;
            screenshotPath: string;
        };

        // Store the metrics in the database
        await drizzle
            .insert(schema.shopSitespeed)
            .values({
                shopId: shop.id,
                createdAt: new Date(),
                ttfb: result.ttfb || null,
                fullyLoaded: result.fullyLoaded || null,
                largestContentfulPaint: result.largestContentfulPaint || null,
                firstContentfulPaint: result.firstContentfulPaint || null,
                cumulativeLayoutShift: result.cumulativeLayoutShift || null,
                transferSize: result.transferSize || null,
            })
            .execute();

        await drizzle
            .update(schema.shop)
            .set({
                shopImage: `/sitespeed/${shop.id}/${result.screenshotPath}`,
            })
            .where(eq(schema.shop.id, shop.id))
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
