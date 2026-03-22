import type { Drizzle } from "#src/db.ts";
import AdminRepository, {
  type OrganizationListParams,
  type ShopListParams,
} from "./admin.repository.ts";

export const listOrganizations = async (db: Drizzle, params?: OrganizationListParams) => {
  return await AdminRepository.listOrganizations(db, params);
};

export const listShops = async (db: Drizzle, params?: ShopListParams) => {
  return await AdminRepository.listShops(db, params);
};

export const getStats = async (db: Drizzle) => {
  return await AdminRepository.getStats(db);
};

export const getGrowthData = async (db: Drizzle) => {
  return await AdminRepository.getGrowthData(db);
};

export const getRecentActivity = async (db: Drizzle) => {
  return await AdminRepository.getRecentActivity(db);
};

export const getShopwareVersions = async (db: Drizzle) => {
  return await AdminRepository.getShopwareVersions(db);
};
