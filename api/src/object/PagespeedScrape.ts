import { Connection } from "@planetscale/database/dist";
import { getConnection } from "../db";
import Shops from "../repository/shops";
import { createSentry } from "../sentry";

interface SQLShop {
    id: number;
    name: string;
    url: string;
    team_id: number;
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

        const id = await this.state.storage.get('id')

        // ID is missing, so we can't do anything
        if (id === undefined) {
            await this.state.storage.deleteAll();
            return;
        }

        const fetchShopSQL = 'SELECT id, name, url, team_id FROM shop WHERE id = ?';

        const shops = await con.execute(fetchShopSQL, [id]);

        // Shop is missing, so we can't do anything
        if (shops.rows.length === 0) {
            console.log(`cannot find shop: ${id}. Destroy self`)
            await this.state.storage.deleteAll();
            return;
        }

        try {
            await this.computePagespeed(shops.rows[0] as SQLShop, con);
        } catch (e) {
            const sentry = createSentry(this.state, this.env);

            sentry.setExtra('shopId', id);
            sentry.captureException(e);
        }

        await this.state.storage.setAlarm(Date.now() + 24 * 60 * MINUTES);
        console.log(`Set alarm for shop ${id} to one day`)
    }

    async computePagespeed(shop: SQLShop, con: Connection) {
        try {
            const home = await fetch(shop.url);

            if (home.status !== 200) {
                await Shops.notify(
                    con,
                    this.env.USER_SOCKET,
                    shop.id.toString(),
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
                shop.id.toString(),
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
            // That can happen that the shop is in maintaince mode and the pagespeed API returns a 503
            return;
        }

        const pagespeed = await pagespeedResponse.json() as PagespeedResponse;

        await con.execute('INSERT INTO shop_pagespeed(shop_id, performance, accessibility, bestpractices, seo) VALUES(?, ?, ?, ?, ?)', [
            shop.id,
            pagespeed.lighthouseResult.categories.performance.score * 100,
            pagespeed.lighthouseResult.categories.accessibility.score * 100,
            pagespeed.lighthouseResult.categories['best-practices'].score * 100,
            pagespeed.lighthouseResult.categories.seo.score * 100,
        ]);
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
    }
}

interface Audit {
    id: string
    title: string
    description: string
    score: number
    manualDescription: string
}