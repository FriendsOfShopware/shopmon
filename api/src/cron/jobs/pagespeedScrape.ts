import { promises as fs } from 'node:fs';
import path from 'node:path';
import { eq } from 'drizzle-orm';
import sharp from 'sharp';
import { type Drizzle, getConnection, schema } from '../../db.ts';
import Shops from '../../repository/shops.ts';

interface SQLShop {
    id: number;
    name: string;
    url: string;
    organizationId: string;
    organizationSlug: string;
    shopImage: string | null;
}

interface PagespeedResponse {
    lighthouseResult: {
        categories: {
            performance: Audit;
            accessibility: Audit;
            'best-practices': Audit;
            seo: Audit;
        };
        audits: {
            'final-screenshot': {
                details: {
                    data: string;
                };
            };
        };
    };
}

interface Audit {
    id: string;
    title: string;
    description: string;
    score: number;
    manualDescription: string;
}

export async function pagespeedScrapeJob() {
    const con = getConnection();

    // Get all shops that need PageSpeed analysis
    const shops = await con.query.shop.findMany({
        columns: {
            id: true,
            name: true,
            url: true,
            organizationId: true,
            shopImage: true,
        },
    });

    console.log(`Found ${shops.length} shops for PageSpeed analysis`);

    // Process shops in parallel with a limit
    const batchSize = 5; // Lower batch size for PageSpeed API
    for (let i = 0; i < shops.length; i += batchSize) {
        const batch = shops.slice(i, i + batchSize);
        await Promise.all(
            batch.map((shop) => computePagespeed(shop as SQLShop, con)),
        );
    }
}

export async function scrapeSinglePagespeedShop(shopId: number) {
    const con = getConnection();

    const shop = await con
        .select({
            id: schema.shop.id,
            name: schema.shop.name,
            url: schema.shop.url,
            organizationId: schema.shop.organizationId,
            organizationSlug: schema.organization.slug,
            shopImage: schema.shop.shopImage,
        })
        .from(schema.shop)
        .innerJoin(
            schema.organization,
            eq(schema.organization.id, schema.shop.organizationId),
        )
        .where(eq(schema.shop.id, shopId))
        .get();

    if (!shop) {
        console.error(`Shop with ID ${shopId} not found`);
        return;
    }

    console.log(`Scraping PageSpeed for shop ${shop.name} (${shop.url})`);
    await computePagespeed(shop as SQLShop, con);
}

async function computePagespeed(shop: SQLShop, con: Drizzle) {
    try {
        const home = await fetch(shop.url);

        if (home.status !== 200) {
            await Shops.notify(
                con,
                shop.id,
                `pagespeed.wrong.status.${shop.id}`,
                {
                    level: 'error',
                    title: `Shop: ${shop.name} could not be checked for Pagespeed`,
                    message: `Could not run Pagespeed against Shop as the http status code is ${home.status}, but expected is 200`,
                    link: {
                        name: 'account.shops.detail',
                        params: {
                            shopId: shop.id.toString(),
                            slug: shop.organizationSlug,
                        },
                    },
                },
            );
            return;
        }
    } catch (e) {
        await Shops.notify(
            con,
            shop.id,
            `pagespeed.shop.not.available.${shop.id}`,
            {
                level: 'error',
                title: `Shop: ${shop.name} could not be checked for Pagespeed`,
                message: `Could not connect to shop. Please check your remote server and try again. Error: ${e}`,
                link: {
                    name: 'account.shops.detail',
                    params: {
                        shopId: shop.id.toString(),
                        slug: shop.organizationSlug,
                    },
                },
            },
        );
        return;
    }

    const params = new URL(
        'https://pagespeedonline.googleapis.com/pagespeedonline/v5/runPagespeed',
    );
    params.searchParams.set('strategy', 'MOBILE');
    params.searchParams.set('url', shop.url);

    if (process.env.PAGESPEED_API_KEY) {
        params.searchParams.set('key', process.env.PAGESPEED_API_KEY);
    }

    params.searchParams.append('category', 'ACCESSIBILITY');
    params.searchParams.append('category', 'BEST_PRACTICES');
    params.searchParams.append('category', 'PERFORMANCE');
    params.searchParams.append('category', 'SEO');

    const pagespeedResponse = await fetch(params.toString());

    if (pagespeedResponse.status !== 200) {
        // That can happen that the shop is in maintenance mode and the PageSpeed API returns a 503
        console.log(
            `PageSpeed API returned ${pagespeedResponse.status} for shop ${shop.id}`,
        );
        return;
    }

    const pagespeed = (await pagespeedResponse.json()) as PagespeedResponse;

    let pageScreenshot =
        pagespeed.lighthouseResult.audits['final-screenshot'].details.data;
    pageScreenshot = pageScreenshot.substr(pageScreenshot.indexOf(',') + 1);

    // Create local files directory if it doesn't exist
    const filesDir = process.env.APP_FILES_DIR || './files';
    const screenshotDir = path.join(filesDir, 'pagespeed', crypto.randomUUID());
    await fs.mkdir(screenshotDir, { recursive: true });

    const fileName = path.join(screenshotDir, 'screenshot.jpg');
    const relativeFileName = path.relative(filesDir, fileName);

    // Convert base64 to buffer and save with sharp for optimization
    const buffer = Buffer.from(pageScreenshot, 'base64');
    await sharp(buffer).jpeg({ quality: 85 }).toFile(fileName);

    // Delete the previous image
    if (shop.shopImage) {
        const oldFile = path.join(filesDir, shop.shopImage);
        try {
            await fs.unlink(oldFile);
            // Also try to remove the parent directory if empty
            const oldDir = path.dirname(oldFile);
            await fs.rmdir(oldDir).catch(() => {}); // Ignore if directory is not empty
        } catch (e) {
            console.error(`Failed to delete old image: ${e}`);
        }
    }

    await con
        .insert(schema.shopPageSpeed)
        .values({
            shopId: shop.id,
            performance:
                pagespeed.lighthouseResult.categories.performance.score * 100,
            accessibility:
                pagespeed.lighthouseResult.categories.accessibility.score * 100,
            bestPractices:
                pagespeed.lighthouseResult.categories['best-practices'].score *
                100,
            seo: pagespeed.lighthouseResult.categories.seo.score * 100,
            createdAt: new Date(),
        })
        .execute();

    await con
        .update(schema.shop)
        .set({ shopImage: relativeFileName })
        .where(eq(schema.shop.id, shop.id))
        .execute();

    console.log(`Updated PageSpeed for shop ${shop.id}`);
}
