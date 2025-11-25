import { eq } from 'drizzle-orm';
import { getScreenshotUrl, runSitespeedReport } from '#src/service/sitespeed';
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
            connectionIssueCount: true,
        },
        where: eq(schema.shop.id, shopId),
    });

    if (!shop) {
        console.error(`Shop ${shopId} not found`);
        return;
    }

    if (
        !shop.sitespeedEnabled ||
        !shop.sitespeedUrls ||
        shop.connectionIssueCount > 0
    ) {
        if (!shop.sitespeedUrls) {
            console.warn(
                `Shop ${shop.id} (${shop.name}) has sitespeed disabled or no URLs configured`,
            );
        }

        if (shop.connectionIssueCount > 0) {
            console.warn(
                `Shop ${shop.id} (${shop.name}) has connection issues, skipping sitespeed analysis`,
            );
        }

        return;
    }

    console.log(`Running sitespeed analysis for shop ${shop.id}: ${shop.name}`);

    shop.sitespeedUrls = shop.sitespeedUrls.map((url) =>
        url.replace('http://localhost:3889', 'http://demoshop:8000'),
    );

    try {
        const result = await runSitespeedReport(shop.id, shop.sitespeedUrls);

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
                shopImage: getScreenshotUrl(shop.id),
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
