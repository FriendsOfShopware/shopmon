import { ref } from "vue";

import { api } from "@/helpers/api";
import { i18n } from "@/i18n";
import { useAlert } from "@/composables/useAlert";
import type { components } from "@/types/api";

type NotificationPreference = components["schemas"]["NotificationPreference"];
type NotificationEventType = components["schemas"]["NotificationEventType"];

export type TriState = "inherit" | "on" | "off";
export type ChannelName = "in_app" | "email";

export const channelList: { channel: ChannelName; labelKey: string; shortLabelKey: string }[] = [
  {
    channel: "in_app",
    labelKey: "settings.channelInApp",
    shortLabelKey: "settings.channelInAppShort",
  },
  {
    channel: "email",
    labelKey: "settings.channelEmail",
    shortLabelKey: "settings.channelEmailShort",
  },
];

export const triStateOptions: { value: TriState; labelKey: string }[] = [
  { value: "inherit", labelKey: "settings.channelInherit" },
  { value: "on", labelKey: "settings.channelOn" },
  { value: "off", labelKey: "settings.channelOff" },
];

// Shared across all consumers (settings page, environment modal) so the
// preference state stays consistent and is not refetched per component.
const preferences = ref<NotificationPreference[]>([]);
const eventTypes = ref<NotificationEventType[]>([]);

/**
 * useNotificationPreferences centralizes reading and writing notification
 * preferences. It mirrors the server resolver's model: a preference is
 * identified by scope (global / environment) x event type x channel. Channel
 * controls are booleans; event-type and per-environment controls are tri-state
 * (inherit / on / off), where "inherit" means "no row, fall back to the less
 * specific scope".
 */
export function useNotificationPreferences() {
  const { error: showError } = useAlert();

  async function loadPreferences() {
    const { data } = await api.GET("/account/notification-preferences");
    preferences.value = data ?? [];
  }

  async function loadEventTypes() {
    const { data } = await api.GET("/notifications/event-types");
    eventTypes.value = data ?? [];
  }

  async function loadAll() {
    await Promise.all([loadPreferences(), loadEventTypes()]);
  }

  function eventTypeLabel(eventType: string): string {
    const key = `settings.eventType.${eventType}`;
    return i18n.global.te(key) ? i18n.global.t(key) : eventType;
  }

  function reportFirstError(errs: unknown[]) {
    const err = errs.find(Boolean);
    if (err) {
      showError((err as { message?: string }).message ?? "");
    }
  }

  // ── Global event x channel (boolean per cell of the matrix) ──
  // Resolves like the server: an event-specific global row wins, else the
  // channel wildcard row, else enabled by default. Toggling writes an explicit
  // event-specific row so each cell is independently controllable.
  function eventChannelEnabled(eventType: string, channel: ChannelName): boolean {
    const specific = preferences.value.find(
      (p) => p.scopeType === "global" && p.eventType === eventType && p.channel === channel,
    );
    if (specific) return specific.enabled;
    const wildcard = preferences.value.find(
      (p) => p.scopeType === "global" && p.eventType === "" && p.channel === channel,
    );
    if (wildcard) return wildcard.enabled;
    return true;
  }

  async function setEventChannel(eventType: string, channel: ChannelName, enabled: boolean) {
    const { error } = await api.PUT("/account/notification-preferences", {
      body: { scopeType: "global", scopeId: "", eventType, channel, enabled },
    });
    reportFirstError([error]);
    await loadPreferences();
  }

  // ── Per-environment event x channel (tri-state per matrix cell) ──
  // The environment row is the most specific scope, so "inherit" (no row) falls
  // back to the global cell for the same event x channel.
  function envEventChannelState(
    environmentId: number,
    eventType: string,
    channel: ChannelName,
  ): TriState {
    const row = preferences.value.find(
      (p) =>
        p.scopeType === "environment" &&
        p.scopeId === String(environmentId) &&
        p.eventType === eventType &&
        p.channel === channel,
    );
    if (!row) return "inherit";
    return row.enabled ? "on" : "off";
  }

  async function setEnvEventChannel(
    environmentId: number,
    eventType: string,
    channel: ChannelName,
    state: TriState,
  ) {
    const scopeId = String(environmentId);
    const { error } =
      state === "inherit"
        ? await api.DELETE("/account/notification-preferences", {
            params: { query: { scopeType: "environment", scopeId, eventType, channel } },
          })
        : await api.PUT("/account/notification-preferences", {
            body: {
              scopeType: "environment",
              scopeId,
              eventType,
              channel,
              enabled: state === "on",
            },
          });
    reportFirstError([error]);
    await loadPreferences();
  }

  return {
    preferences,
    eventTypes,
    channelList,
    triStateOptions,
    loadAll,
    loadPreferences,
    loadEventTypes,
    eventTypeLabel,
    eventChannelEnabled,
    setEventChannel,
    envEventChannelState,
    setEnvEventChannel,
  };
}
