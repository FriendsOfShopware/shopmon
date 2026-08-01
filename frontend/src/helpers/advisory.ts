// Shared severity styling for the advisory list and detail pages.

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
  if (score == null) return "muted";
  if (score >= 7) return "destructive";
  if (score >= 4) return "warning";
  return "muted";
}
