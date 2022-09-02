export interface Shop {
    id: number,
    status: 'green' | 'yellow' | 'red',
    name: string,
    url: string,
    shopware_version: string,
    team_id: number,
    team_name: string,
    last_scraped_at: string;
    last_scraped_error: string;
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
    latestVersion: string|null,
    installed: boolean,
    ratingAverage: number|null,
    storeLink: string|null,
    changelog: ExtensionChangelog[]|null,
}

export interface ExtensionChangelog {
    version: string
    text: string
    creationDate: string
}

export interface ScheduledTask {
    name: string,
    status: string,
    latestVersion: string,
    lastExecutionTime: string,
    nextExecutionTime: string,
}