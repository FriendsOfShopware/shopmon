import "#src/sentry.ts";
import { Worker } from "bullmq";
import { getRedisConnection } from "#src/modules/queue/connection.ts";
import { getShopQueue, getSitespeedQueue, getMaintenanceQueue } from "#src/modules/queue/queues.ts";
import { shopScrapeJob, scrapeSingleShop } from "#src/modules/shop/jobs/shop-scrape.job.ts";
import {
  scrapeSitespeedForAllShops,
  scrapeSingleSitespeedShop,
} from "#src/modules/shop/jobs/sitespeed-scrape.job.ts";
import { lockCleanupJob } from "#src/modules/shared/jobs/lock-cleanup.job.ts";
import { invitationCleanupJob } from "#src/modules/organization/jobs/invitation-cleanup.job.ts";

const connection = getRedisConnection();

// Shop queue worker — handles scrape jobs
const shopWorker = new Worker(
  "shop",
  async (job) => {
    switch (job.name) {
      case "scrape-all":
        await shopScrapeJob();
        break;
      case "scrape":
        await scrapeSingleShop(job.data.shopId);
        break;
      default:
        console.warn(`Unknown shop job: ${job.name}`);
    }
  },
  {
    connection,
    concurrency: 10,
  },
);

// Sitespeed queue worker — isolated with low concurrency
const sitespeedWorker = new Worker(
  "sitespeed",
  async (job) => {
    switch (job.name) {
      case "sitespeed-all":
        await scrapeSitespeedForAllShops();
        break;
      case "sitespeed":
        await scrapeSingleSitespeedShop(job.data.shopId);
        break;
      default:
        console.warn(`Unknown sitespeed job: ${job.name}`);
    }
  },
  {
    connection,
    concurrency: 1,
  },
);

// Maintenance queue worker — handles cleanup jobs
const maintenanceWorker = new Worker(
  "maintenance",
  async (job) => {
    switch (job.name) {
      case "lock-cleanup":
        await lockCleanupJob();
        break;
      case "invitation-cleanup":
        await invitationCleanupJob();
        break;
      default:
        console.warn(`Unknown maintenance job: ${job.name}`);
    }
  },
  {
    connection,
    concurrency: 1,
  },
);

// Error handlers
for (const worker of [shopWorker, sitespeedWorker, maintenanceWorker]) {
  worker.on("failed", (job, err) => {
    console.error(`Job ${job?.name} (${job?.id}) failed:`, err);
  });
}

// Register repeatable job schedulers
async function registerRepeatableJobs() {
  const shopQueue = getShopQueue();
  const sitespeedQueue = getSitespeedQueue();
  const maintenanceQueue = getMaintenanceQueue();

  // Shop scrape every hour
  await shopQueue.upsertJobScheduler(
    "scrape-all-hourly",
    { pattern: "0 * * * *" },
    { name: "scrape-all" },
  );

  // Sitespeed scrape daily at 3 AM
  await sitespeedQueue.upsertJobScheduler(
    "sitespeed-all-daily",
    { pattern: "0 3 * * *" },
    { name: "sitespeed-all" },
  );

  // Lock cleanup daily at 4 AM
  await maintenanceQueue.upsertJobScheduler(
    "lock-cleanup-daily",
    { pattern: "0 4 * * *" },
    { name: "lock-cleanup" },
  );

  // Invitation cleanup daily at 5 AM
  await maintenanceQueue.upsertJobScheduler(
    "invitation-cleanup-daily",
    { pattern: "0 5 * * *" },
    { name: "invitation-cleanup" },
  );

  console.log("Registered repeatable job schedulers");
}

registerRepeatableJobs().then(() => {
  console.log("Workers started, waiting for jobs...");
});

// Graceful shutdown
function shutdown() {
  console.log("Shutting down workers...");
  Promise.all([shopWorker.close(), sitespeedWorker.close(), maintenanceWorker.close()]).then(() => {
    process.exit(0);
  });
}

process.on("SIGTERM", shutdown);
process.on("SIGINT", shutdown);
