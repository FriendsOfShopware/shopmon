import type { Team } from './team'

export enum SHOP_STATUS {
    GREEN = 'green',
    YELLOW = 'yellow',
    RED = 'red',
}

export interface lastUpdated {
    date?: string;
    from?: string;
    to?: string;
}

export interface Shop {
    id: number;
    status: SHOP_STATUS;
    name: string;
    favicon: string | null;
    url: string;
    shopware_version: string;
    last_updated: lastUpdated | null;
    team_id: Team['id'];
    team_name: Team['name'];
    last_scraped_at: string;
    last_scraped_error: string;
}

export interface ShopDetailed extends Shop {
    extensions: Extension[];
    scheduled_task: ScheduledTask[];
    cache_info: CacheInfo;
    queue_info: QueueInfo[];
    pagespeed: Pagespeed[];
    changelog: ShopChangelog[];
    shop_image: string;
    checks: CheckerChecks[];
    ignores: string[];
}

export interface CheckerChecks {
    id: string;
    level: SHOP_STATUS;
    message: string;
    source: string;
    link: string | null;
}

export interface Extension {
    name: string,
    label: string,
    active: boolean,
    version: string,
    latestVersion: string | null,
    installed: boolean,
    ratingAverage: number | null,
    storeLink: string | null,
    changelog: ExtensionChangelog[] | null,
    installedAt: string | null,
}

export interface UserExtension extends Extension {
    shops: {
        [key: string]: {
            id: number,
            name: string,
            team_id: number,
            shopware_version: string,
            installed: boolean,
            active: boolean,
            version: string
        }
    }
}

export interface ExtensionDiff {
    name: string,
    label: string,
    state: string,
    old_version: string | null,
    new_version: string | null,
    changelog: ExtensionChangelog[] | null,
    active: boolean,
}

export interface ExtensionChangelog {
    version: string
    text: string
    creationDate: string
    isCompatible: boolean;
}

export interface ExtensionCompatibilitys extends Extension {
    compatibility: {
        label: string,
        name: string,
        type: string
    } | null
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

export interface ShopChangelog {
    id: number;
    shop_id: number;
    extensions: ExtensionDiff[];
    old_shopware_version: string | null;
    new_shopware_version: string | null;
    date: string;
}

export interface ShopwareVersion {
    [version: string]: string[];
}
