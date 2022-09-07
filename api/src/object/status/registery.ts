import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { CacheInfo, Extension, QueueInfo, ScheduledTask, SHOP_STATUS } from "../../../../shared/shop"; 
import env from "./checks/env";
import frosh_tools from "./checks/frosh_tools";
import task from "./checks/task";
import worker from "./checks/worker";

export interface CheckerInput {
    extensions: Extension[],
    config: {
        version: string;
        adminWorker: {
            enableAdminWorker: boolean;
        }
    };
    scheduledTasks: ScheduledTask[];
    queueInfo: QueueInfo[];
    cacheInfo: CacheInfo;
    favicon: string|null;
    client: HttpClient;
}

export interface CheckerChecks {
    level: SHOP_STATUS;
    message: string;
    source: string;
    link: string|null;
}

export class CheckerOutput {
    public status: SHOP_STATUS = SHOP_STATUS.GREEN;
    public checks: CheckerChecks[] = [];

    public success(message: string, source = 'Shopmon', link: string|null = null) {
        this.checks.push({
            level: SHOP_STATUS.GREEN,
            message: message,
            source: source,
            link
        });
    }

    public warning(message: string, source = 'Shopmon', link: string|null = null) {
        this.checks.push({
            level: SHOP_STATUS.YELLOW,
            message: message,
            source: source,
            link
        });

        if (this.status === SHOP_STATUS.GREEN) {
            this.status = SHOP_STATUS.YELLOW;
        }
    }

    public error(message: string, source = 'Shopmon', link: string|null = null) {
        this.checks.push({
            level: SHOP_STATUS.RED,
            message: message,
            source: source,
            link
        });

        this.status = SHOP_STATUS.RED;
    }
}

export interface Checker {
    check(input: CheckerInput, result: CheckerOutput): Promise<void>;
}

export class CheckerRegistery {
    static async check(input: CheckerInput) {
        const result = new CheckerOutput();

        await Promise.all([
            new env().check(input, result),
            new task().check(input, result),
            new worker().check(input, result),
            new frosh_tools().check(input, result),
        ]);

        return result;
    }
}