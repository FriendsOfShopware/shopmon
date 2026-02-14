import { eq, inArray } from "drizzle-orm";
import {
  type Drizzle,
  shopExtension,
  storeExtension,
  storeExtensionVersion,
} from "#src/db.ts";
import type { Extension } from "#src/types/index.ts";
import versionCompare from "#src/util.ts";
import UsersRepository from "./user.repository.ts";

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
  const shops = await UsersRepository.findShopsSimple(db, userId);

  if (shops.length === 0) {
    return [];
  }

  const shopIds = shops.map((s) => s.id);
  const shopMap = new Map(shops.map((s) => [s.id, s]));

  // Query all extensions for all user's shops in one query, joining store data
  const extensions = await db
    .select({
      shopId: shopExtension.shopId,
      name: shopExtension.name,
      label: shopExtension.label,
      active: shopExtension.active,
      version: shopExtension.version,
      latestVersion: shopExtension.latestVersion,
      installed: shopExtension.installed,
      installedAt: shopExtension.installedAt,
      storeExtensionId: shopExtension.storeExtensionId,
      ratingAverage: storeExtension.ratingAverage,
      storeLink: storeExtension.storeLink,
    })
    .from(shopExtension)
    .leftJoin(storeExtension, eq(shopExtension.storeExtensionId, storeExtension.id))
    .where(inArray(shopExtension.shopId, shopIds));

  // Collect all store extension IDs for changelog lookup
  const storeExtIds = [
    ...new Set(
      extensions.map((ext) => ext.storeExtensionId).filter((id): id is number => id !== null),
    ),
  ];

  const allVersions =
    storeExtIds.length > 0
      ? await db
          .select()
          .from(storeExtensionVersion)
          .where(inArray(storeExtensionVersion.storeExtensionId, storeExtIds))
      : [];

  const versionsByExtId = new Map<number, typeof allVersions>();
  for (const v of allVersions) {
    const list = versionsByExtId.get(v.storeExtensionId) || [];
    list.push(v);
    versionsByExtId.set(v.storeExtensionId, list);
  }

  const json = {} as { [key: string]: UserExtension };

  for (const ext of extensions) {
    const shop = shopMap.get(ext.shopId);
    if (!shop) continue;

    if (json[ext.name] === undefined) {
      let changelog: Extension["changelog"] = null;

      if (ext.storeExtensionId && ext.latestVersion && ext.latestVersion !== ext.version) {
        const versions = versionsByExtId.get(ext.storeExtensionId) || [];
        const entries = versions
          .filter((v) => versionCompare(v.version, ext.version) > 0)
          .map((v) => ({
            version: v.version,
            text: v.changelog || "",
            creationDate: v.releaseDate || "",
            isCompatible: versionCompare(v.version, ext.latestVersion!) <= 0,
          }));
        if (entries.length > 0) changelog = entries;
      }

      json[ext.name] = {
        name: ext.name,
        label: ext.label,
        active: ext.active,
        version: ext.version,
        latestVersion: ext.latestVersion,
        installed: ext.installed,
        ratingAverage: ext.ratingAverage ?? null,
        storeLink: ext.storeLink ?? null,
        changelog,
        installedAt: ext.installedAt,
        shops: {},
      } as UserExtension;
    }

    json[ext.name].shops[ext.shopId] = {
      id: shop.id,
      name: shop.name,
      organizationId: shop.organizationId,
      organizationSlug: shop.organizationSlug,
      shopwareVersion: shop.shopwareVersion,
      installed: ext.installed,
      active: ext.active,
      version: ext.version,
    };
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

export const getSubscribedShops = async (db: Drizzle, userId: string, notifications: string[]) => {
  if (notifications.length === 0) {
    return [];
  }

  // Extract shop IDs from notifications array
  const shopIds = notifications
    .filter((n) => n.startsWith("shop-"))
    .map((n) => Number.parseInt(n.replace("shop-", ""), 10))
    .filter((id) => !Number.isNaN(id));

  if (shopIds.length === 0) {
    return [];
  }

  return await UsersRepository.findSubscribedShops(db, userId, shopIds);
};
