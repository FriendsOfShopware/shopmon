import { auth } from "#src/auth.ts";
import { closeConnection, getConnection, schema } from "#src/db.ts";
import { eq } from "drizzle-orm";

const skipShop = process.argv.includes("--skip-shop");

async function main() {
  const user1 = await auth.api.signUpEmail({
    body: {
      email: "owner@fos.gg",
      password: "password",
      name: "Owner",
    },
  });

  const org = await auth.api.createOrganization({
    body: {
      name: "Acme Corp",
      slug: "acme-corp",
      userId: user1.user.id,
    },
  });

  const user2 = await auth.api.signUpEmail({
    body: {
      email: "admin@fos.gg",
      password: "password",
      name: "Admin",
    },
  });

  await auth.api.addMember({
    body: {
      role: "admin",
      userId: user2.user.id,
      organizationId: org.id,
    },
  });

  const user3 = await auth.api.signUpEmail({
    body: {
      email: "member@fos.gg",
      password: "password",
      name: "Member",
    },
  });

  await auth.api.addMember({
    body: {
      role: "member",
      userId: user3.user.id,
      organizationId: org.id,
    },
  });

  const user4 = await auth.api.signUpEmail({
    body: {
      email: "regular@fos.gg",
      password: "password",
      name: "Regular",
    },
  });

  await getConnection().update(schema.user).set({ emailVerified: true }).execute();

  await getConnection()
    .update(schema.user)
    .set({ role: "admin" })
    .where(eq(schema.user.id, user1.user.id))
    .execute();

  if (!skipShop) {
    const { HttpClient } = await import("#src/modules/shop/http-client.ts");
    const { scrapeSingleShop } = await import("#src/modules/shop/jobs/shop-scrape.job.ts");
    const { encrypt } = await import("#src/modules/shop/crypto.ts");
    const { default: shops, generateShopToken } =
      await import("#src/modules/shop/shop.repository.ts");
    const { S3Client } = await import("bun");

    await getConnection().insert(schema.project).values({
      id: 1,
      name: "Acme Shop",
      organizationId: org.id,
      createdAt: new Date(),
      updatedAt: new Date(),
    });

    const shopUrl = "http://localhost:3889";
    const shopClientId = "SWIAUZL4OXRKEG1RR3PMCEVNMG";
    const shopClientSecret = "aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks";
    const shopToken = generateShopToken();

    const client = new HttpClient({
      url: shopUrl,
      clientId: shopClientId,
      clientSecret: shopClientSecret,
      shopToken,
    });

    const resp = await client.get<{ version: string }>("/_info/config");

    const shopId = await shops.createShop(getConnection(), {
      name: "Local",
      organizationId: org.id,
      projectId: 1,
      shopUrl,
      clientId: shopClientId,
      clientSecret: await encrypt(process.env.APP_SECRET, shopClientSecret),
      version: resp.body.version,
      shopToken,
    });

    await scrapeSingleShop(shopId);

    // Create fake deployments over the last 7 days
    const s3 = new S3Client({
      endpoint: process.env.APP_S3_ENDPOINT,
      accessKeyId: process.env.APP_S3_ACCESS_KEY_ID,
      secretAccessKey: process.env.APP_S3_SECRET_ACCESS_KEY,
      bucket: process.env.APP_S3_BUCKET,
      region: process.env.APP_S3_REGION,
    });

    const deploymentCommands = [
      "bin/console app:install",
      "bin/console theme:compile",
      "bin/console cache:clear",
      "bin/console dal:refresh:index",
      "bin/console plugin:update",
      "bin/console scheduled-task:run",
      "bin/console messenger:consume",
      "bin/console assets:install",
    ];

    const deploymentNames = [
      "v6.5.8.0 Release",
      "Hotfix Login",
      "Theme Update",
      "Plugin Refresh",
      "Index Rebuild",
      "Cache Warmup",
      "Messenger Fix",
      "Asset Deploy",
    ];

    const deploymentLogs = [
      `Installing app "SwagPayPal"...\nApp "SwagPayPal" has been successfully installed.\nClearing cache...\nCache cleared successfully.\n`,
      `Compiling theme for sales channel "Storefront"...\nProcessing SCSS files...\nProcessing JS files...\nTheme compiled successfully in 34.2s.\n`,
      `Clearing all caches...\n  - Clearing HTTP cache... done\n  - Clearing object cache... done\n  - Clearing template cache... done\nAll caches cleared successfully.\n`,
      `Refreshing index for "product"...\nRefreshing index for "category"...\nERROR: Elasticsearch connection refused at localhost:9200\nIndex refresh failed with exit code 1.\n`,
      `Updating plugin "SwagPayPal" from 7.1.0 to 7.2.0...\nMigrating database...\nRunning migration "Migration1699999999UpdatePayPalSettings"...\nPlugin updated successfully.\n`,
      `Running scheduled tasks...\n  - product_export.generate: executed in 2.3s\n  - log_entry.cleanup: executed in 0.4s\n  - newsletter_recipient.cleanup: executed in 0.1s\nAll scheduled tasks completed.\n`,
      `Consuming messages from "async" transport...\nProcessed 142 messages in 45.3s.\n  - ProductIndexingMessage: 89\n  - CategoryIndexingMessage: 31\n  - MediaIndexingMessage: 22\nConsumer finished.\n`,
      `Installing assets for all bundles...\n  - Storefront: 234 files copied\n  - Administration: 567 files copied\n  - SwagPayPal: 12 files copied\nAssets installed successfully.\n`,
    ];

    const now = new Date();
    const deploymentIds: { id: number; createdAt: Date }[] = [];

    for (let i = 0; i < 8; i++) {
      const hoursAgo = Math.floor((7 * 24 * i) / 8); // spread over 7 days
      const startDate = new Date(now.getTime() - hoursAgo * 60 * 60 * 1000);
      const executionTime = Math.random() * 120 + 5; // 5-125 seconds
      const endDate = new Date(startDate.getTime() + executionTime * 1000);
      const returnCode = i === 3 ? 1 : 0; // one failed deployment

      const [inserted] = await getConnection()
        .insert(schema.deployment)
        .values({
          shopId,
          name: deploymentNames[i],
          command: deploymentCommands[i],
          returnCode,
          startDate,
          endDate,
          executionTime,
          composer: { "shopware/core": "6.5.8.0", "shopware/storefront": "6.5.8.0" },
          reference: `abc${(1000 + i).toString()}`,
          createdAt: startDate,
        })
        .returning({ id: schema.deployment.id });

      deploymentIds.push({ id: inserted.id, createdAt: startDate });

      const compressed = Bun.zstdCompressSync(Buffer.from(deploymentLogs[i]));
      await s3.file(`deployments/${inserted.id}/output.zst`).write(compressed);
    }

    console.log("Created 8 fake deployments with S3 logs over the last 7 days");

    // Create fake sitespeed entries over the last 7 days
    // Generate entries every ~6 hours, some linked to deployments
    const sitespeedEntries: (typeof schema.shopSitespeed.$inferInsert)[] = [];
    const baseTtfb = 120;
    const baseFullyLoaded = 2800;
    const baseLcp = 1800;
    const baseFcp = 900;
    const baseCls = 0.05;
    const baseTransferSize = 1_200_000;

    for (let hoursAgo = 7 * 24; hoursAgo >= 0; hoursAgo -= 6) {
      const createdAt = new Date(now.getTime() - hoursAgo * 60 * 60 * 1000);

      // Find if a deployment happened close to this time (within 2 hours after)
      const linkedDeployment = deploymentIds.find((d) => {
        const diff = createdAt.getTime() - d.createdAt.getTime();
        return diff >= 0 && diff < 2 * 60 * 60 * 1000;
      });

      // Add some variance; deployments cause a temporary performance dip
      const deploymentPenalty = linkedDeployment ? 1.3 : 1;
      const jitter = () => 0.85 + Math.random() * 0.3; // 0.85 - 1.15

      sitespeedEntries.push({
        shopId,
        deploymentId: linkedDeployment?.id ?? null,
        createdAt,
        ttfb: Math.round(baseTtfb * jitter() * deploymentPenalty),
        fullyLoaded: Math.round(baseFullyLoaded * jitter() * deploymentPenalty),
        largestContentfulPaint: Math.round(baseLcp * jitter() * deploymentPenalty),
        firstContentfulPaint: Math.round(baseFcp * jitter() * deploymentPenalty),
        cumulativeLayoutShift: Math.round(baseCls * jitter() * deploymentPenalty * 1000) / 1000,
        transferSize: Math.round(baseTransferSize * jitter()),
      });
    }

    await getConnection().insert(schema.shopSitespeed).values(sitespeedEntries);

    console.log(`Created ${sitespeedEntries.length} fake sitespeed entries over the last 7 days`);
    console.log("Shop:", shopUrl);
  } else {
    console.log("Skipping shop creation (--skip-shop)");
  }

  console.log("Fixtures applied successfully");
  console.log("User 1 (org owner):", user1.user.email);
  console.log("User 2 (org admin):", user2.user.email);
  console.log("User 3 (org member):", user3.user.email);
  console.log("User 4 (regular user without organization):", user4.user.email);
  console.log("Organization:", org.name);

  console.log('All users have the password "password".');

  await closeConnection();
}

main();
