import { Queue } from "bullmq";
import { getRedisConnection } from "./connection.ts";

let shopQueue: Queue | undefined;
let sitespeedQueue: Queue | undefined;
let maintenanceQueue: Queue | undefined;

export function getShopQueue(): Queue {
  if (!shopQueue) {
    shopQueue = new Queue("shop", { connection: getRedisConnection() });
  }
  return shopQueue;
}

export function getSitespeedQueue(): Queue {
  if (!sitespeedQueue) {
    sitespeedQueue = new Queue("sitespeed", { connection: getRedisConnection() });
  }
  return sitespeedQueue;
}

export function getMaintenanceQueue(): Queue {
  if (!maintenanceQueue) {
    maintenanceQueue = new Queue("maintenance", { connection: getRedisConnection() });
  }
  return maintenanceQueue;
}

export async function addShopScrapeJob(shopId: number) {
  const queue = getShopQueue();
  return queue.add("scrape", { shopId });
}

export async function addShopSitespeedJob(shopId: number, opts?: { delay?: number }) {
  const queue = getSitespeedQueue();
  return queue.add("sitespeed", { shopId }, opts);
}
