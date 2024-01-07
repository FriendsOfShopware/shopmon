import { Drizzle, getConnection, schema } from "../db";
import Shops from "../repository/shops";
import { createSentry } from "../toucan";
import { eq } from 'drizzle-orm';

interface SQLShop {
    id: number;
    name: string;
    url: string;
    team_id: number;
    shop_image: string | null;
}

const SECONDS = 1000;
const MINUTES = 60 * SECONDS;

export class PagespeedScrape implements DurableObject {
    state: DurableObjectState;
    env: Env;

    constructor(state: DurableObjectState, env: Env) {
        this.env = env;
        this.state = state;
    }

    async fetch(request: Request): Promise<Response> {
        const url = new URL(request.url);
        const id = url.searchParams.get('id')

        await this.state.storage.put('id', id)

        if (url.pathname === '/cron') {
            const currentAlarm = await this.state.storage.getAlarm();
            if (currentAlarm === null) {
                await this.state.storage.setAlarm(Date.now() + 24 * 60 * MINUTES);
                console.log(`Set alarm for shop ${id} to one day`)
            }

            return new Response('OK');
        } else if (url.pathname === '/now') {
            await this.state.storage.setAlarm(Date.now() + 5 * SECONDS);
            console.log(`Set alarm for shop ${id} to 5 seconds`)

            return new Response('OK');
        } else if (url.pathname === '/delete') {
            await this.state.storage.deleteAll();

            return new Response('OK');
        }

        return new Response('', { status: 404 });
    }

    async alarm(): Promise<void> {
        const con = getConnection(this.env);

        const id = await this.state.storage.get('id') as string | undefined;

        // ID is missing, so we can't do anything
        if (id === undefined) {
            await this.state.storage.deleteAll();
            return;
        }

        const shop = await con.query.shop.findFirst({
            columns: {
                id: true,
                name: true,
                url: true,
                team_id: true,
            },
            where: eq(schema.shop.id, parseInt(id))
        })

        // Shop is missing, so we can't do anything
        if (shop === undefined) {
            console.log(`cannot find shop: ${id}. Destroy self`)
            await this.state.storage.deleteAll();
            return;
        }

        try {
            await this.computePagespeed(shop as SQLShop, con);
        } catch (e) {
            const sentry = createSentry(this.state, this.env);

            sentry.setExtra('shopId', id);
            sentry.captureException(e);
        }

        await this.state.storage.setAlarm(Date.now() + 24 * 60 * MINUTES);
        console.log(`Set alarm for shop ${id} to one day`)
    }

    async computePagespeed(shop: SQLShop, con: Drizzle) {
        try {
            const home = await fetch(shop.url);

            if (home.status !== 200) {
                await Shops.notify(
                    con,
                    this.env.USER_SOCKET,
                    shop.id,
                    `pagespeed.wrong.status.${shop.id}`,
                    {
                        level: 'error',
                        title: `Shop: ${shop.name} could not be checked for Pagespeed`,
                        message: `Could not run Pagespeed against Shop as the http status code is ${home.status}, but expected is 200`,
                        link: { name: 'account.shops.detail', params: { shopId: shop.id.toString(), teamId: shop.team_id.toString() } }
                    }
                )
                return;
            }

        } catch (e) {
            await Shops.notify(
                con,
                this.env.USER_SOCKET,
                shop.id,
                `pagespeed.shop.not.available.${shop.id}`,
                {
                    level: 'error',
                    title: `Shop: ${shop.name} could not be checked for Pagespeed`,
                    message: `Could not connect to shop. Please check your remote server and try again. Error: ${e}`,
                    link: { name: 'account.shops.detail', params: { shopId: shop.id.toString(), teamId: shop.team_id.toString() } }
                }
            )
        }

        const params = new URL('https://pagespeedonline.googleapis.com/pagespeedonline/v5/runPagespeed');
        params.searchParams.set('strategy', 'MOBILE');
        params.searchParams.set('url', shop.url);
        params.searchParams.set('key', this.env.PAGESPEED_API_KEY);
        params.searchParams.append('category', 'ACCESSIBILITY');
        params.searchParams.append('category', 'BEST_PRACTICES');
        params.searchParams.append('category', 'PERFORMANCE');
        params.searchParams.append('category', 'SEO');

        const pagespeedResponse = await fetch(params.toString());

        if (pagespeedResponse.status !== 200) {
            // That can happen that the shop is in maintenance mode and the PageSpeed API returns a 503
            return;
        }

        const pagespeed = await pagespeedResponse.json() as PagespeedResponse;

        let pageScreenshot = pagespeed.lighthouseResult.audits['final-screenshot'].details.data;
        pageScreenshot = pageScreenshot.substr(pageScreenshot.indexOf(',') + 1);

        const fileName = `pagespeed/${crypto.randomUUID()}/screenshot.jpg`;

        await this.env.FILES.put(fileName, base64ToArrayBuffer(pageScreenshot));

        // Delete the previous image
        if (shop.shop_image) {
            await this.env.FILES.delete(shop.shop_image);
        }

        await con
            .insert(schema.shopPageSpeed)
            .values({
                shop_id: shop.id,
                performance: pagespeed.lighthouseResult.categories.performance.score * 100,
                accessibility: pagespeed.lighthouseResult.categories.accessibility.score * 100,
                best_practices: pagespeed.lighthouseResult.categories['best-practices'].score * 100,
                seo: pagespeed.lighthouseResult.categories.seo.score * 100,
                created_at: new Date().toISOString(),
            }).execute();


        await con.update(schema.shop).set({ shop_image: fileName }).where(eq(schema.shop.id, shop.id)).execute();
    }
}

interface PagespeedResponse {
    lighthouseResult: {
        categories: {
            performance: Audit
            accessibility: Audit
            "best-practices": Audit
            seo: Audit
        }
        audits: {
            "final-screenshot": {
                details: {
                    data: string;
                }
            }
        }
    }
}

interface Audit {
    id: string
    title: string
    description: string
    score: number
    manualDescription: string
}

function base64ToArrayBuffer(base64: string) {
    const binary_string = atob(base64);
    const len = binary_string.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
        bytes[i] = binary_string.charCodeAt(i);
    }
    return bytes.buffer;
}
