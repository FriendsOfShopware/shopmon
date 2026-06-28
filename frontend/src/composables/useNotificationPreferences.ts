import { ref } from "vue";

import { api } from "@/helpers/api";
import { i18n } from "@/i18n";
import { useAlert } from "@/composables/useAlert";
import type { components } from "@/types/api";

type NotificationPreference = components["schemas"]["NotificationPreference"];
type NotificationEventType = components["schemas"]["NotificationEventType"];

export type TriState = "inherit" | "on" | "off";
export type ChannelName = "in_app" | "email";

export const channelList: { channel: ChannelName; labelKey: string }[] = [
  { channel: "in_app", labelKey: "settings.channelInApp" },
  { channel: "email", labelKey: "settings.channelEmail" },
];

export const triStateOptions: { value: TriState; labelKey: string }[] = [
  { value: "inherit", labelKey: "settings.channelInherit" },
  { value: "on", labelKey: "settings.channelOn" },
  { value: "off", labelKey: "settings.channelOff" },
];

const CHANNELS: ChannelName[] = ["in_app", "email"];

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

  // ── Global channel (boolean: enabled unless a global wildcard disable row) ──
  function globalChannelEnabled(channel: ChannelName): boolean {
    return !preferences.value.some(
      (p) => p.scopeType === "global" && p.eventType === "" && p.channel === channel && !p.enabled,
    );
  }

  async function setGlobalChannel(channel: ChannelName, enabled: boolean) {
    const { error } = await api.PUT("/account/notification-preferences", {
      body: { scopeType: "global", scopeId: "", eventType: "", channel, enabled },
    });
    reportFirstError([error]);
    await loadPreferences();
  }

  // ── Global event type (boolean across both channels) ──
  function eventTypeEnabled(eventType: string): boolean {
    return !preferences.value.some(
      (p) =>
        p.scopeType === "global" &&
        p.eventType === eventType &&
        (p.channel === "in_app" || p.channel === "email") &&
        !p.enabled,
    );
  }

  async function setEventType(eventType: string, enabled: boolean) {
    const results = await Promise.all(
      CHANNELS.map((channel) =>
        enabled
          ? api.DELETE("/account/notification-preferences", {
              params: { query: { scopeType: "global", scopeId: "", eventType, channel } },
            })
          : api.PUT("/account/notification-preferences", {
              body: { scopeType: "global", scopeId: "", eventType, channel, enabled: false },
            }),
      ),
    );
    reportFirstError(results.map((r) => r.error));
    await loadPreferences();
  }

  // ── Per-environment channel (tri-state) ──
  function envChannelState(environmentId: number, channel: ChannelName): TriState {
    const row = preferences.value.find(
      (p) =>
        p.scopeType === "environment" &&
        p.scopeId === String(environmentId) &&
        p.eventType === "" &&
        p.channel === channel,
    );
    if (!row) return "inherit";
    return row.enabled ? "on" : "off";
  }

  async function setEnvChannel(environmentId: number, channel: ChannelName, state: TriState) {
    const scopeId = String(environmentId);
    const { error } =
      state === "inherit"
        ? await api.DELETE("/account/notification-preferences", {
            params: { query: { scopeType: "environment", scopeId, eventType: "", channel } },
          })
        : await api.PUT("/account/notification-preferences", {
            body: {
              scopeType: "environment",
              scopeId,
              eventType: "",
              channel,
              enabled: state === "on",
            },
          });
    reportFirstError([error]);
    await loadPreferences();
  }

  // ── Per-environment event type (tri-state across both channels) ──
  function envEventState(environmentId: number, eventType: string): TriState {
    const rows = preferences.value.filter(
      (p) =>
        p.scopeType === "environment" &&
        p.scopeId === String(environmentId) &&
        p.eventType === eventType &&
        (p.channel === "in_app" || p.channel === "email"),
    );
    if (rows.length === 0) return "inherit";
    return rows.some((p) => !p.enabled) ? "off" : "on";
  }

  async function setEnvEvent(environmentId: number, eventType: string, state: TriState) {
    const scopeId = String(environmentId);
    const results = await Promise.all(
      CHANNELS.map((channel) =>
        state === "inherit"
          ? api.DELETE("/account/notification-preferences", {
              params: { query: { scopeType: "environment", scopeId, eventType, channel } },
            })
          : api.PUT("/account/notification-preferences", {
              body: {
                scopeType: "environment",
                scopeId,
                eventType,
                channel,
                enabled: state === "on",
              },
            }),
      ),
    );
    reportFirstError(results.map((r) => r.error));
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
    globalChannelEnabled,
    setGlobalChannel,
    eventTypeEnabled,
    setEventType,
    envChannelState,
    setEnvChannel,
    envEventState,
    setEnvEvent,
  };
}
