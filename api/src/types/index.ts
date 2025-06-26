export type Extension = {
    name: string;
    label: string;
    active: boolean;
    version: string;
    latestVersion: string | null;
    installed: boolean;
    ratingAverage: number | null;
    storeLink: string | null;
    changelog: ExtensionChangelog[] | null;
    installedAt: string | null;
};

export type ExtensionChangelog = {
    version: string;
    text: string;
    creationDate: string;
    isCompatible: boolean;
};

export type ScheduledTask = {
    id: string;
    name: string;
    status: string;
    interval: number;
    overdue: boolean;
    lastExecutionTime: string;
    nextExecutionTime: string;
};

export type QueueInfo = {
    name: string;
    size: number;
};

export type CacheInfo = {
    environment: string;
    httpCache: boolean;
    cacheAdapter: string;
};

export type ExtensionDiff = {
    name: string;
    label: string;
    state: string;
    old_version: string | null;
    new_version: string | null;
    changelog: ExtensionChangelog[] | null;
    active: boolean;
};

export type ShopStatus = 'green' | 'yellow' | 'red';

export type CheckerChecks = {
    id: string;
    level: ShopStatus;
    message: string;
    source: string;
    link: string | null;
};

export type NotificationLink = {
    name: string;
    params: Record<string, string>;
} | null;

export type ShopScrapeInfo = {
    extensions: Extension[];
    scheduledTask: ScheduledTask[];
    queueInfo: QueueInfo[];
    cacheInfo: CacheInfo;
    checks: CheckerChecks[];
    createdAt: Date;
};
