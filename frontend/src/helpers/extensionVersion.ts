import { compareVersions, validate } from "compare-versions";

// Unknown versions cannot establish that a release is an upgrade.
export function isNewerExtensionVersion(installed: string, available?: string | null): boolean {
  if (!available || !validate(installed) || !validate(available)) return false;
  return compareVersions(available, installed) > 0;
}
