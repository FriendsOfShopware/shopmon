import type { NotificationPreference } from "@/api/generated";
export function globalEventEnabled(
  preferences: NotificationPreference[],
  eventType: string,
  channel: string,
): boolean {
  const specific = preferences.find(
    (p) => p.scopeType === "global" && p.eventType === eventType && p.channel === channel,
  );
  if (specific) return specific.enabled;
  return (
    preferences.find((p) => p.scopeType === "global" && p.eventType === "" && p.channel === channel)
      ?.enabled ?? true
  );
}
export function environmentEventState(
  preferences: NotificationPreference[],
  environmentId: number,
  eventType: string,
  channel: string,
): "inherit" | "on" | "off" {
  const row = preferences.find(
    (p) =>
      p.scopeType === "environment" &&
      p.scopeId === String(environmentId) &&
      p.eventType === eventType &&
      p.channel === channel,
  );
  return row ? (row.enabled ? "on" : "off") : "inherit";
}
