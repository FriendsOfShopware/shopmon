import type { Drizzle } from '#src/db.ts';
import { getShopScrapeInfo } from '#src/modules/shop/scrape-info.repository.ts';
import UsersRepository from './user.repository.ts';

interface Extension {
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
}

interface ExtensionChangelog {
    version: string;
    text: string;
    creationDate: string;
    isCompatible: boolean;
}

export interface UserExtension extends Extension {
    shops: {
        [key: string]: {
            id: number;
            name: string;
            organizationId: string;
            organizationSlug: string;
            shopwareVersion: string;
            installed: boolean;
            active: boolean;
            version: string;
        };
    };
}

export const getCurrentUser = async (db: Drizzle, userId: string) => {
    return await UsersRepository.findById(db, userId);
};

export const getCurrentUserExtensions = async (db: Drizzle, userId: string) => {
    const result = await UsersRepository.findShopsSimple(db, userId);

    const json = {} as { [key: string]: UserExtension };

    for (const row of result) {
        const scrapeResult = await getShopScrapeInfo(row.id);

        if (!scrapeResult) {
            continue;
        }

        for (const extension of scrapeResult.extensions) {
            if (json[extension.name] === undefined) {
                json[extension.name] = extension as UserExtension;
                json[extension.name].shops = {};
            }

            json[extension.name].shops[row.id] = {
                id: row.id,
                name: row.name,
                organizationId: row.organizationId,
                organizationSlug: row.organizationSlug,
                shopwareVersion: row.shopwareVersion,
                installed: extension.installed,
                active: extension.active,
                version: extension.version,
            };
        }
    }

    return Object.values(json);
};

export const listOrganizations = async (db: Drizzle, userId: string) => {
    return await UsersRepository.findOrganizations(db, userId);
};

export const getCurrentUserShops = async (db: Drizzle, userId: string) => {
    return await UsersRepository.findShops(db, userId);
};

export const getCurrentUserProjects = async (db: Drizzle, userId: string) => {
    return await UsersRepository.findProjects(db, userId);
};

export const getCurrentUserChangelogs = async (db: Drizzle, userId: string) => {
    return await UsersRepository.findChangelogs(db, userId);
};

export const getSubscribedShops = async (
    db: Drizzle,
    userId: string,
    notifications: string[],
) => {
    if (notifications.length === 0) {
        return [];
    }

    // Extract shop IDs from notifications array
    const shopIds = notifications
        .filter((n) => n.startsWith('shop-'))
        .map((n) => Number.parseInt(n.replace('shop-', ''), 10))
        .filter((id) => !Number.isNaN(id));

    if (shopIds.length === 0) {
        return [];
    }

    return await UsersRepository.findSubscribedShops(db, userId, shopIds);
};
