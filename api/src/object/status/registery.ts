import { CacheInfo, Extension, QueueInfo, ScheduledTask, SHOP_STATUS } from "../../../../shared/shop"; 
import env from "./checks/env";
import task from "./checks/task";

export interface CheckerInput {
    extensions: Extension[],
    scheduledTasks: ScheduledTask[];
    queueInfo: QueueInfo[];
    cacheInfo: CacheInfo;
    favicon: string|null;
}

export interface CheckerChecks {
    level: SHOP_STATUS;
    message: string;
    source: string;
}

export class CheckerOutput {
    public status: SHOP_STATUS = SHOP_STATUS.GREEN;
    public checks: CheckerChecks[] = [];

    public success(message: string, source: string = 'Shopmon') {
        this.checks.push({
            level: SHOP_STATUS.GREEN,
            message: message,
            source: source,
        });
    }

    public warning(message: string, source: string = 'Shopmon') {
        this.checks.push({
            level: SHOP_STATUS.YELLOW,
            message: message,
            source: source,
        });

        if (this.status === SHOP_STATUS.GREEN) {
            this.status = SHOP_STATUS.YELLOW;
        }
    }

    public error(message: string, source: string = 'Shopmon') {
        this.checks.push({
            level: SHOP_STATUS.RED,
            message: message,
            source: source,
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
        ]);

        return result;
    }
}