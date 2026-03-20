import { and, eq, isNull } from "drizzle-orm";
import { getConnection, schema, shopExtension } from "#src/db.ts";
import { fetchComposerRepoVersions } from "#src/modules/shop/composer-repo.service.ts";
import versionCompare from "#src/util.ts";

export async function composerCheckJob(shopId: number) {
  const con = getConnection();

  // Read shop's composer repositories
  const shopData = await con.query.shop.findFirst({
    columns: { composerRepositories: true },
    where: eq(schema.shop.id, shopId),
  });

  if (!shopData?.composerRepositories?.length) {
    return;
  }

  // Read extensions still missing a latestVersion
  const unresolvedExtensions = await con
    .select({
      id: shopExtension.id,
      name: shopExtension.name,
      version: shopExtension.version,
    })
    .from(shopExtension)
    .where(and(eq(shopExtension.shopId, shopId), isNull(shopExtension.latestVersion)));

  if (unresolvedExtensions.length === 0) {
    return;
  }

  for (const repo of shopData.composerRepositories) {
    try {
      const repoVersions = await fetchComposerRepoVersions(repo);

      for (const ext of unresolvedExtensions) {
        const repoResult = repoVersions.get(ext.name);
        if (!repoResult) continue;

        // Build changelog from version history
        let changelog = null;
        if (repoResult.latestVersion !== ext.version) {
          changelog = repoResult.versions
            .filter((v) => versionCompare(v.version, ext.version) > 0)
            .map((v) => ({
              version: v.version,
              text: "",
              creationDate: v.time ?? new Date().toISOString(),
              isCompatible: versionCompare(v.version, repoResult.latestVersion) <= 0,
            }))
            .sort((a, b) => versionCompare(b.version, a.version));
        }

        await con
          .update(shopExtension)
          .set({
            latestVersion: repoResult.latestVersion,
            changelog,
          })
          .where(eq(shopExtension.id, ext.id))
          .execute();
      }
    } catch (e) {
      console.warn(`Failed to check custom repo ${repo.url}: ${e}`);
    }
  }

  console.log(`Composer check completed for shop ${shopId}`);
}
