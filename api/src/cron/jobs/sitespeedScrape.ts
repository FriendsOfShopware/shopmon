import { eq } from 'drizzle-orm';
import { schema } from '../../db.ts';
import type { Drizzle } from '../../db.ts';

export async function scrapeSitespeedForAllShops() {
    const { getConnection } = await import('../../db.ts');
    const drizzle = getConnection();

    console.log('Starting sitespeed scrape for all shops');

    const shops = await drizzle.query.shop.findMany({
        columns: {
            id: true,
            url: true,
            name: true,
        },
    });

    console.log(`Found ${shops.length} shops to analyze`);

    const sitespeedServiceUrl =
        process.env.APP_SITESPEED_ENDPOINT || 'http://localhost:3001';

    for (const shop of shops) {
        try {
            console.log(
                `Running sitespeed analysis for shop ${shop.id}: ${shop.name}`,
            );

            const response = await fetch(`${sitespeedServiceUrl}/analyze`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    shopId: shop.id,
                    url: shop.url,
                }),
            });

            if (!response.ok) {
                console.error(
                    `Sitespeed service error for shop ${shop.id}: ${response.statusText}`,
                );
                continue;
            }

            const result = await response.json();

            // Store the metrics in the database
            await drizzle
                .insert(schema.shopSitespeed)
                .values({
                    shopId: shop.id,
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
                `Error running sitespeed analysis for shop ${shop.id}:`,
                error,
            );
        }
    }

    console.log('Sitespeed scrape completed for all shops');
}

export async function scrapeSingleSitespeedShop(shopId: number) {
    const { getConnection } = await import('../../db.ts');
    const drizzle = getConnection();

    const shop = await drizzle.query.shop.findFirst({
        columns: {
            id: true,
            url: true,
            name: true,
        },
        where: eq(schema.shop.id, shopId),
    });

    if (!shop) {
        console.error(`Shop ${shopId} not found`);
        return;
    }

    console.log(`Running sitespeed analysis for shop ${shop.id}: ${shop.name}`);

    const sitespeedServiceUrl =
        process.env.APP_SITESPEED_ENDPOINT || 'http://localhost:3001';

    try {
        const response = await fetch(`${sitespeedServiceUrl}/analyze`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                shopId: shop.id,
                url: shop.url,
            }),
        });

        if (!response.ok) {
            throw new Error(`Sitespeed service error: ${response.statusText}`);
        }

        const result = await response.json();

        // Store the metrics in the database
        await drizzle
            .insert(schema.shopSitespeed)
            .values({
                shopId: shop.id,
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
            `Error running sitespeed analysis for shop ${shop.id}:`,
            error,
        );
        throw error;
    }
}
