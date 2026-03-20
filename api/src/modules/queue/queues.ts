import { Queue } from "bullmq";
import { getRedisConnection } from "./connection.ts";

let shopQueue: Queue | undefined;
let sitespeedQueue: Queue | undefined;
let maintenanceQueue: Queue | undefined;
let composerQueue: Queue | undefined;

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

export function getComposerQueue(): Queue {
  if (!composerQueue) {
    composerQueue = new Queue("composer", { connection: getRedisConnection() });
  }
  return composerQueue;
}

export async function addShopScrapeJob(shopId: number) {
  const queue = getShopQueue();
  return queue.add("scrape", { shopId });
}

export async function addShopSitespeedJob(shopId: number, opts?: { delay?: number }) {
  const queue = getSitespeedQueue();
  return queue.add("sitespeed", { shopId }, opts);
}

export async function addComposerCheckJob(shopId: number) {
  const queue = getComposerQueue();
  return queue.add("composer-check", { shopId });
}
