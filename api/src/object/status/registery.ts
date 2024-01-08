import { HttpClient } from '@friendsofshopware/app-server-sdk';
import {
    CacheInfo,
    Extension,
    QueueInfo,
    ScheduledTask,
} from '../../../../frontend/src/types/shop';
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
    extensions: Extension[];
    config: {
        version: string;
        adminWorker: {
            enableAdminWorker: boolean;
        };
    };
    scheduledTasks: ScheduledTask[];
    queueInfo: QueueInfo[];
    cacheInfo: CacheInfo;
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
