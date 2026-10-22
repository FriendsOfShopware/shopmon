// Shared severity styling and link safety for the advisory pages.

import type { AdvisoryCwe } from "@/api/generated";

export function advisoryCwesFromInput(value: string, known: AdvisoryCwe[]): AdvisoryCwe[] {
  const names = new Map(known.map((cwe) => [cwe.id.toUpperCase(), cwe.name]));
  return value
    .split(",")
    .map((id) => id.trim().toUpperCase())
    .filter(Boolean)
    .map((id) => ({ id, name: names.get(id) ?? "" }));
}

type AdvisoryOriginLike = {
  origin?: string | null;
  cve?: string | null;
  ghsaId?: string | null;
};

// Whether the advisory was authored by Shopmon rather than synced from an
// upstream registry. Shopmon reports are badged everywhere they surface.
export function isShopmonAdvisory(advisory: AdvisoryOriginLike): boolean {
  return advisory.origin === "shopmon";
}

// The official identifier a Shopmon advisory was linked to, if any: CVE first,
// then GHSA — the same precedence the backend uses for public ids.
export function advisoryOfficialId(advisory: AdvisoryOriginLike): string | null {
  return advisory.cve || advisory.ghsaId || null;
}

export function isLinkedShopmonAdvisory(advisory: AdvisoryOriginLike): boolean {
  return isShopmonAdvisory(advisory) && advisoryOfficialId(advisory) !== null;
}

// Whether the advisory is tracked manually (authored by Shopmon or imported
// from GitHub) rather than synced from Packagist. Manual advisories expose
// content editing and official linking in the admin UI.
export function isManualAdvisory(advisory: AdvisoryOriginLike): boolean {
  return advisory.origin === "shopmon" || advisory.origin === "github";
}

// Whether the sync owns the advisory's core fields: once an official CVE is
// linked, upstream data wins and content editing closes. A GHSA-only manual
// advisory stays editable.
export function isAdvisoryContentLocked(advisory: AdvisoryOriginLike): boolean {
  return !!advisory.cve;
}

// Advisory URLs come from Packagist, NVD, and admin enrichment — none of which
// Shopmon controls. rel="noopener noreferrer" does not stop a javascript: or
// data: href from executing on click, so only http(s) links are ever rendered
// as clickable.
export function isSafeHttpUrl(url: string | null | undefined): boolean {
  if (!url) return false;
  try {
    const protocol = new URL(url).protocol;
    return protocol === "http:" || protocol === "https:";
  } catch {
    return false;
  }
}

export function severityBadgeClass(level: string | null | undefined): string {
  switch (level) {
    case "critical":
      return "bg-destructive/15 text-destructive border-destructive/20";
    case "high":
      return "bg-orange-500/15 text-orange-700 dark:text-orange-300 border-orange-500/20";
    case "medium":
      return "bg-warning/15 text-warning border-warning/20";
    case "low":
      return "bg-blue-500/15 text-blue-700 dark:text-blue-300 border-blue-500/20";
    default:
      return "bg-muted text-muted-foreground";
  }
}

// Icon tile (like StatCard / ListShops leading icon) tinted by severity.
export function severityTileClass(level: string | null | undefined): string {
  switch (level) {
    case "critical":
    case "high":
      return "bg-destructive/10 text-destructive";
    case "medium":
      return "bg-warning/10 text-warning";
    case "low":
      return "bg-blue-500/10 text-blue-600 dark:text-blue-400";
    default:
      return "bg-muted text-muted-foreground";
  }
}

// StatCard colour for a severity level.
export function severityStatColor(
  level: string | null | undefined,
): "destructive" | "warning" | "muted" {
  switch (level) {
    case "critical":
    case "high":
      return "destructive";
    case "medium":
      return "warning";
    default:
      return "muted";
  }
}

// StatCard colour for a CVSS base score (0-10).
export function cvssStatColor(
  score: number | null | undefined,
): "destructive" | "warning" | "muted" {
  if (score === null || score === undefined) return "muted";
  if (score >= 7) return "destructive";
  if (score >= 4) return "warning";
  return "muted";
}
