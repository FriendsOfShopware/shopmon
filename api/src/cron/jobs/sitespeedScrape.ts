import { eq } from 'drizzle-orm';
import { getConnection, schema } from '../../db.ts';
import { sanitizeSitespeedLabel } from '../../util.ts';

export async function scrapeSitespeedForAllShops() {
    const drizzle = getConnection();

    console.log('Starting sitespeed scrape for all shops');

    const shops = await drizzle.query.shop.findMany({
        columns: {
            id: true,
            url: true,
            name: true,
            sitespeedEnabled: true,
            sitespeedUrls: true,
        },
        where: eq(schema.shop.sitespeedEnabled, true),
    });

    console.log(`Found ${shops.length} shops with sitespeed enabled`);

    const sitespeedServiceUrl =
        process.env.APP_SITESPEED_ENDPOINT || 'http://localhost:3001';

    for (const shop of shops) {
        // Default to shop URL if no specific URLs are configured
        const urlsToTest =
            shop.sitespeedUrls.length > 0
                ? shop.sitespeedUrls.slice(0, 5) // Limit to 5 URLs
                : [{ url: shop.url, label: 'Homepage' }];

        for (const urlConfig of urlsToTest) {
            try {
                const sanitizedFolderName = sanitizeSitespeedLabel(
                    urlConfig.label,
                    urlConfig.url,
                );

                console.log(
                    `Running sitespeed analysis for shop ${shop.id}: ${shop.name} - ${urlConfig.label} (${sanitizedFolderName})`,
                );

                const response = await fetch(`${sitespeedServiceUrl}/analyze`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        shopId: shop.id,
                        url: urlConfig.url,
                        label: urlConfig.label,
                        folderName: sanitizedFolderName,
                    }),
                });

                if (!response.ok) {
                    console.error(
                        `Sitespeed service error for shop ${shop.id} - ${urlConfig.label}: ${response.statusText}`,
                    );
                    continue;
                }

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
                        url: urlConfig.url,
                        label: urlConfig.label,
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

                console.log(
                    `Sitespeed analysis completed for shop ${shop.id} - ${urlConfig.label}`,
                );
            } catch (error) {
                console.error(
                    `Error running sitespeed analysis for shop ${shop.id} - ${urlConfig.label}:`,
                    error,
                );
            }
        }
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

    const sitespeedServiceUrl =
        process.env.APP_SITESPEED_ENDPOINT || 'http://localhost:3001';

    // Default to shop URL if no specific URLs are configured
    const urlsToTest =
        shop.sitespeedUrls.length > 0
            ? shop.sitespeedUrls.slice(0, 5) // Limit to 5 URLs
            : [{ url: shop.url, label: 'Homepage' }];

    for (const urlConfig of urlsToTest) {
        try {
            const sanitizedFolderName = sanitizeSitespeedLabel(
                urlConfig.label,
                urlConfig.url,
            );

            const response = await fetch(`${sitespeedServiceUrl}/analyze`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    shopId: shop.id,
                    url: urlConfig.url,
                    label: urlConfig.label,
                    folderName: sanitizedFolderName,
                }),
            });

            if (!response.ok) {
                throw new Error(
                    `Sitespeed service error: ${response.statusText}`,
                );
            }

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
                    url: urlConfig.url,
                    label: urlConfig.label,
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

            console.log(
                `Sitespeed analysis completed for shop ${shop.id} - ${urlConfig.label}`,
            );
        } catch (error) {
            console.error(
                `Error running sitespeed analysis for shop ${shop.id} - ${urlConfig.label}:`,
                error,
            );
            throw error;
        }
    }
}
