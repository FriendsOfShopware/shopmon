import type { components } from "@/types/api";

export type AccountExtension = components["schemas"]["AccountExtension"];
export type AccountExtensionEnvironment = components["schemas"]["AccountExtensionEnvironment"];
export type ExtensionScreenshot = components["schemas"]["ExtensionScreenshot"];

// Normalizes a possibly protocol-relative producer website URL (the store often
// returns "//example.com") into an absolute https URL safe for an href.
export function normalizeWebsite(url?: string | null): string | null {
  if (!url) return null;
  if (url.startsWith("//")) return `https:${url}`;
  if (!/^https?:\/\//.test(url)) return `https://${url}`;
  return url;
}

// True when the extension is installed and a newer version is available in the store.
export function hasUpdate(ext: AccountExtension): boolean {
  return !!(ext.installed && ext.latestVersion && ext.version !== ext.latestVersion);
}

// True when a specific environment install is running an outdated version.
export function envHasUpdate(env: AccountExtensionEnvironment): boolean {
  return !!(env.installed && env.latestVersion && env.version !== env.latestVersion);
}

// Number of environments that still need to update this extension.
export function updateCount(ext: AccountExtension): number {
  return ext.environments.filter((e) => envHasUpdate(e)).length;
}

export function isInactive(ext: AccountExtension): boolean {
  return ext.installed && !ext.active;
}

export function extensionState(ext: AccountExtension): "not installed" | "active" | "inactive" {
  if (!ext.installed) return "not installed";
  if (ext.active) return "active";
  return "inactive";
}

export function envState(
  env: AccountExtensionEnvironment,
): "not installed" | "active" | "inactive" {
  if (!env.installed) return "not installed";
  if (env.active) return "active";
  return "inactive";
}
