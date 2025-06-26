import type { HttpClient } from '@shopware-ag/app-server-sdk';
import type {
    CacheInfo,
    Extension,
    QueueInfo,
    ScheduledTask,
} from '../../types/index.ts';
import env from './checks/env.ts';
import frosh_tools from './checks/frosh_tools.ts';
import security from './checks/security.ts';
import task from './checks/task.ts';
import worker from './checks/worker.ts';

type SHOP_STATUS = 'green' | 'yellow' | 'red';

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
    status: SHOP_STATUS = 'green';
    checks: CheckerChecks[] = [];
    ignores: string[];

    constructor(ignores: string[]) {
        this.ignores = ignores;
    }

    success(
        id: string,
        message: string,
        source = 'Shopmon',
        link: string | null = null,
    ) {
        this.checks.push({
            id,
            level: 'green',
            message: message,
            source: source,
            link,
        });
    }

    warning(
        id: string,
        message: string,
        source = 'Shopmon',
        link: string | null = null,
    ) {
        this.checks.push({
            id,
            level: 'yellow',
            message: message,
            source: source,
            link,
        });

        if (this.ignores.indexOf(id) !== -1) {
            return;
        }

        this.status = 'yellow';
    }

    error(
        id: string,
        message: string,
        source = 'Shopmon',
        link: string | null = null,
    ) {
        this.checks.push({
            id,
            level: 'red',
            message: message,
            source: source,
            link,
        });

        if (this.ignores.indexOf(id) !== -1) {
            return;
        }

        this.status = 'red';
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
