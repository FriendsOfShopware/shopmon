import { decrypt } from "#src/modules/shop/crypto.ts";
import versionCompare from "#src/util.ts";

interface ComposerPackageVersion {
  version: string;
  time?: string;
  extra?: {
    "shopware-plugin-class"?: string;
  };
}

export interface ComposerRepoResult {
  latestVersion: string;
  versions: { version: string; time?: string }[];
}

interface ComposerPackagesJson {
  packages: Record<string, Record<string, ComposerPackageVersion> | ComposerPackageVersion[]>;
  "metadata-url"?: string;
  "available-packages"?: string[];
}

function extractTechnicalName(pluginClass: string): string {
  const parts = pluginClass.split("\\");
  return parts[parts.length - 1];
}

function isStableVersion(version: string): boolean {
  if (version.startsWith("dev-") || version.startsWith("v-dev")) return false;
  const lower = version.toLowerCase();
  return !lower.includes("alpha") && !lower.includes("beta") && !lower.includes("rc");
}

function processVersionList(
  versions: ComposerPackageVersion[],
  result: Map<string, ComposerRepoResult>,
): void {
  let latestVersion: string | null = null;
  let technicalName: string | null = null;
  const stableVersions: { version: string; time?: string }[] = [];

  for (const meta of versions) {
    if (!isStableVersion(meta.version)) continue;

    const pluginClass = meta.extra?.["shopware-plugin-class"];
    if (!pluginClass) continue;

    technicalName = extractTechnicalName(pluginClass);

    const normalizedVersion = meta.version.replace(/^v/, "");
    stableVersions.push({ version: normalizedVersion, time: meta.time });
    if (!latestVersion || versionCompare(normalizedVersion, latestVersion) > 0) {
      latestVersion = normalizedVersion;
    }
  }

  if (technicalName && latestVersion) {
    result.set(technicalName, { latestVersion, versions: stableVersions });
  }
}

async function fetchPackageMetadata(
  packageName: string,
  metadataUrlTemplate: string,
  baseUrl: string,
  headers: Record<string, string>,
  result: Map<string, ComposerRepoResult>,
): Promise<void> {
  const metadataPath = metadataUrlTemplate.replace("%package%", packageName);
  const metadataUrl = metadataPath.startsWith("http") ? metadataPath : baseUrl + metadataPath;

  const pkgResp = await fetch(metadataUrl, { headers });
  if (!pkgResp.ok) return;

  const pkgData = (await pkgResp.json()) as {
    packages: Record<string, ComposerPackageVersion[]>;
  };

  const versions = pkgData.packages?.[packageName];
  if (Array.isArray(versions)) {
    processVersionList(versions, result);
  }
}

export async function fetchComposerRepoVersions(repo: {
  url: string;
  authType: "none" | "http-basic" | "bearer";
  username?: string;
  password?: string;
  token?: string;
}): Promise<Map<string, ComposerRepoResult>> {
  const technicalNameToLatest = new Map<string, ComposerRepoResult>();

  const headers: Record<string, string> = {};
  if (repo.authType === "http-basic" && repo.username && repo.password) {
    const decryptedPassword = await decrypt(process.env.APP_SECRET, repo.password);
    headers.Authorization = `Basic ${btoa(`${repo.username}:${decryptedPassword}`)}`;
  } else if (repo.authType === "bearer" && repo.token) {
    const decryptedToken = await decrypt(process.env.APP_SECRET, repo.token);
    headers.Authorization = `Bearer ${decryptedToken}`;
  }

  const baseUrl = repo.url.replace(/\/+$/, "");
  const packagesUrl = baseUrl + "/packages.json";
  const resp = await fetch(packagesUrl, { headers });

  if (!resp.ok) {
    console.warn(`Failed to fetch ${packagesUrl}: ${resp.status}`);
    return technicalNameToLatest;
  }

  const data = (await resp.json()) as ComposerPackagesJson;

  // Composer v2 lazy provider: uses metadata-url to fetch individual packages
  if (data["metadata-url"]) {
    const metadataUrlTemplate = data["metadata-url"];

    if (data["available-packages"]?.length) {
      // Fetch only listed packages
      for (const packageName of data["available-packages"]) {
        try {
          await fetchPackageMetadata(
            packageName,
            metadataUrlTemplate,
            baseUrl,
            headers,
            technicalNameToLatest,
          );
        } catch {
          // Skip individual package failures
        }
      }
    } else {
      // available-packages is optional per Composer spec;
      // fall through to process inline packages if present
      for (const [packageName, versions] of Object.entries(data.packages)) {
        if (Array.isArray(versions)) {
          processVersionList(versions, technicalNameToLatest);
        } else {
          // Try fetching via metadata-url for each package in the inline list
          try {
            await fetchPackageMetadata(
              packageName,
              metadataUrlTemplate,
              baseUrl,
              headers,
              technicalNameToLatest,
            );
          } catch {
            // Skip individual package failures
          }
        }
      }
    }

    return technicalNameToLatest;
  }

  // Composer v1 inline format: all versions in packages object
  if (data.packages && Object.keys(data.packages).length > 0) {
    for (const [, versions] of Object.entries(data.packages)) {
      if (Array.isArray(versions)) {
        processVersionList(versions, technicalNameToLatest);
      } else {
        // v1 dict format: { "1.0.0": {...}, "2.0.0": {...} }
        const versionList = Object.entries(versions).map(([versionString, meta]) => ({
          ...meta,
          version: versionString,
        }));
        processVersionList(versionList, technicalNameToLatest);
      }
    }
  }

  return technicalNameToLatest;
}
