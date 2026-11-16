import { ref } from "vue";
import {
  getDestinationNotificationPreferences,
  setDestinationNotificationPreference,
  deleteDestinationNotificationPreference,
  type NotificationPreference,
} from "@/api/generated";
import { globalEventEnabled, environmentEventState } from "@/helpers/notificationPreferences";
import { useNotificationPreferences, type TriState } from "./useNotificationPreferences";
import { useAlert } from "./useAlert";
import { useI18n } from "vue-i18n";

export type PreferenceDestination = { id: string; organizationId: string };
export function useDestinationNotificationPreferences(
  destination: () => PreferenceDestination | undefined,
  environmentId: () => number | undefined,
) {
  const preferences = ref<NotificationPreference[]>([]);
  const loading = ref(false);
  const failed = ref(false);
  const busy = ref(false);
  const shared = useNotificationPreferences();
  const alert = useAlert();
  const { t } = useI18n();
  let generation = 0;
  const path = () => ({ orgId: destination()!.organizationId, destinationId: destination()!.id });
  async function load() {
    if (!destination()) return;
    const current = ++generation;
    loading.value = true;
    failed.value = false;
    try {
      const [response, eventsLoaded] = await Promise.all([
        getDestinationNotificationPreferences({ path: path() }),
        shared.loadEventTypes(),
      ]);
      if (current !== generation) return;
      if (response.error || !response.data || !eventsLoaded) {
        failed.value = true;
        return;
      }
      preferences.value = response.data;
    } catch {
      if (current === generation) failed.value = true;
    } finally {
      if (current === generation) loading.value = false;
    }
  }
  function enabled(eventType: string) {
    return globalEventEnabled(preferences.value, eventType, "destination");
  }
  function state(eventType: string): TriState {
    return environmentEventState(preferences.value, environmentId()!, eventType, "destination");
  }
  async function set(eventType: string, value: TriState | boolean) {
    if (!destination() || busy.value || failed.value || loading.value) return;
    busy.value = true;
    const preference = {
      scopeType: environmentId() == null ? "global" : "environment",
      scopeId: environmentId() == null ? "" : String(environmentId()),
      eventType,
      channel: "destination",
    };
    try {
      const result =
        value === "inherit"
          ? await deleteDestinationNotificationPreference({ path: path(), query: preference })
          : await setDestinationNotificationPreference({
              path: path(),
              body: { ...preference, enabled: value === true || value === "on" },
            });
      if (result.error) {
        alert.error(t("destinations.failedPreferences"));
        return;
      }
      await load();
    } catch {
      alert.error(t("destinations.failedPreferences"));
    } finally {
      busy.value = false;
    }
  }
  return { loading, failed, busy, load, enabled, state, set };
}
