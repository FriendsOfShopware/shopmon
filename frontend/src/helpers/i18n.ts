import { i18n } from "@/i18n";
import type { components } from "@/types/api";

type Params = Record<string, unknown>;

type Notification = components["schemas"]["Notification"];
type EnvironmentCheck = components["schemas"]["EnvironmentCheck"];
type StatusReason = components["schemas"]["StatusReason"];

function tx(key: string, params: Params): string {
  return i18n.global.t(key, params);
}

function hasKey(key: string): boolean {
  return i18n.global.te(key);
}

function isNonEmpty(value: unknown): value is string {
  return typeof value === "string" && value.length > 0;
}

/**
 * Renders a notification title, preferring the translation key (rendered in the
 * viewer's locale) and falling back to the server-rendered title for legacy
 * rows that predate key-based notifications.
 */
export function translateNotificationTitle(n: Notification): string {
  if (n.titleKey && hasKey(n.titleKey)) {
    return tx(n.titleKey, (n.params ?? {}) as Params);
  }
  return n.title;
}

/** Renders a notification message; see translateNotificationTitle. */
export function translateNotificationMessage(n: Notification): string {
  if (n.messageKey && hasKey(n.messageKey)) {
    return tx(n.messageKey, (n.params ?? {}) as Params);
  }
  return n.message;
}

/**
 * Renders a keyed message in the viewer's locale. Mirrors the server RenderCheck
 * logic: translate the key, degrade to the raw `snippet` param when the key is
 * unknown, fall back to the provided text, and append the recommendation clause
 * when current/recommended are present.
 */
function renderKeyed(
  messageKey: string | null | undefined,
  params: Params,
  fallback: string,
): string {
  let msg = fallback;
  if (messageKey && hasKey(messageKey)) {
    msg = tx(messageKey, params);
  } else if (isNonEmpty(params.snippet)) {
    msg = params.snippet;
  }

  if (
    isNonEmpty(params.current) &&
    isNonEmpty(params.recommended) &&
    hasKey("check.recommendationSuffix")
  ) {
    msg += tx("check.recommendationSuffix", params);
  }

  return msg;
}

/** Renders an environment check message in the viewer's locale. */
export function translateCheckMessage(check: EnvironmentCheck): string {
  return renderKeyed(check.messageKey, (check.params ?? {}) as Params, check.message ?? "");
}

/** Renders a status-transition reason (a changed check) in the viewer's locale. */
export function translateReason(reason: StatusReason): string {
  return renderKeyed(reason.messageKey, (reason.params ?? {}) as Params, "");
}

/**
 * Extracts the status-transition reasons embedded in a notification's params
 * (set for status change/recovery events), or an empty list otherwise.
 */
export function notificationReasons(n: Notification): StatusReason[] {
  const reasons = (n.params as { reasons?: unknown } | undefined)?.reasons;
  return Array.isArray(reasons) ? (reasons as StatusReason[]) : [];
}
