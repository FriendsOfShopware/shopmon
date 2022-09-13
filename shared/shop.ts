import type { Team } from './team'

export enum SHOP_STATUS {
    GREEN = 'green',
    YELLOW = 'yellow',
    RED = 'red',
}

export interface Shop {
    id: number,
    status: SHOP_STATUS,
    name: string,
    favicon: string|null;
    url: string,
    shopware_version: string,
    team_id: Team['id'],
    team_name: Team['name'],
    last_scraped_at: string;
    last_scraped_error: string;
}

export interface ShopDetailed extends Shop {
    extensions: Extension[];
    scheduled_task: ScheduledTask[];
    cache_info: CacheInfo;
    queue_info: QueueInfo[];
    pagespeed: Pagespeed[];
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
    installedAt: string|null,
}

export interface ExtensionChangelog {
    version: string
    text: string
    creationDate: string
    isCompatible: boolean;
}

export interface ScheduledTask {
    name: string,
    status: string,
    interval: number,
    overdue: boolean;
    lastExecutionTime: string,
    nextExecutionTime: string,
}

export interface QueueInfo {
    name: string;
    size: number;
}

export interface CacheInfo {
    environment: string;
    httpCache: boolean;
    cacheAdapter: string;
}

export interface Pagespeed {
    id: number;
    shop_id: number;
    created_at: string;
    performance: number;
    accessibility: number;
    best_practices: number;
    seo: number;
}