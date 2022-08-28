export interface Shop {
    id: number,
    name: string,
    url: string,
    shopware_version: string,
    team_id: number,
    team_name: string,
    last_scraped_at: string;
}

export interface ShopDetailed extends Shop {
    extensions: Extension[];
    scheduled_task: ScheduledTask[];
    cache_info: {
        environment: string;
        httpCache: boolean;
        cacheAdapter: string;
    };
}

export interface Extension {
    name: string,
    active: boolean,
    version: string,
    latestVersion: string,
    installed: boolean,
}

export interface ScheduledTask {
    name: string,
    status: string,
    latestVersion: string,
    lastExecutionTime: string,
    nextExecutionTime: string,
}