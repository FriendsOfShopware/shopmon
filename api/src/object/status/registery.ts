import type { HttpClient } from '@shopware-ag/app-server-sdk';
import type { schema } from '../../db';
import env from './checks/env';
import frosh_tools from './checks/frosh_tools';
import security from './checks/security';
import task from './checks/task';
import worker from './checks/worker';

enum SHOP_STATUS {
    GREEN = 'green',
    YELLOW = 'yellow',
    RED = 'red',
}

export interface CheckerInput {
    extensions: typeof schema.shopScrapeInfo.$inferInsert.extensions;
    config: {
        version: string;
        adminWorker: {
            enableAdminWorker: boolean;
        };
    };
    scheduledTasks: typeof schema.shopScrapeInfo.$inferInsert.scheduledTask;
    queueInfo: typeof schema.shopScrapeInfo.$inferInsert.queueInfo;
    cacheInfo: typeof schema.shopScrapeInfo.$inferInsert.cacheInfo;
    favicon: string | null;
    client: HttpClient;
    ignores: string[];
}

export interface CheckerChecks {
    id: string;
    level: SHOP_STATUS;
    message: string;
    source: string;
    link: string | null;
}

export class CheckerOutput {
    public status: SHOP_STATUS = SHOP_STATUS.GREEN;
    public checks: CheckerChecks[] = [];
    ignores: string[];

    constructor(ignores: string[]) {
        this.ignores = ignores;
    }

    public success(
        id: string,
        message: string,
        source = 'Shopmon',
        link: string | null = null,
    ) {
        this.checks.push({
            id,
            level: SHOP_STATUS.GREEN,
            message: message,
            source: source,
            link,
        });
    }

    public warning(
        id: string,
        message: string,
        source = 'Shopmon',
        link: string | null = null,
    ) {
        this.checks.push({
            id,
            level: SHOP_STATUS.YELLOW,
            message: message,
            source: source,
            link,
        });

        if (this.ignores.indexOf(id) !== -1) {
            return;
        }

        this.status = SHOP_STATUS.YELLOW;
    }

    public error(
        id: string,
        message: string,
        source = 'Shopmon',
        link: string | null = null,
    ) {
        this.checks.push({
            id,
            level: SHOP_STATUS.RED,
            message: message,
            source: source,
            link,
        });

        if (this.ignores.indexOf(id) !== -1) {
            return;
        }

        this.status = SHOP_STATUS.RED;
    }
}

export interface Checker {
    check(input: CheckerInput, result: CheckerOutput): Promise<void>;
}

export async function check(input: CheckerInput) {
    const result = new CheckerOutput(input.ignores);

    await Promise.all([
        new security().check(input, result),
        new env().check(input, result),
        new task().check(input, result),
        new worker().check(input, result),
        new frosh_tools().check(input, result),
    ]);

    return result;
}
