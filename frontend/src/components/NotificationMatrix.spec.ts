import { beforeEach, expect, it, vi } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import NotificationMatrix from "./NotificationMatrix.vue";
import { useNotificationPreferences } from "@/composables/useNotificationPreferences";
const alerts = vi.hoisted(() => ({ error: vi.fn() }));
vi.mock("@/composables/useAlert", () => ({ useAlert: () => alerts }));
vi.mock("@/api/generated", () => ({
  getNotificationEventTypes: vi.fn(),
  getNotificationPreferences: vi.fn(),
  setNotificationPreference: vi.fn(),
  deleteNotificationPreference: vi.fn(),
  getDestinationNotificationPreferences: vi.fn(),
  setDestinationNotificationPreference: vi.fn(),
  deleteDestinationNotificationPreference: vi.fn(),
}));
import * as api from "@/api/generated";
const destination = { id: "dest-1", organizationId: "org" };
const globalOff = {
  scopeType: "global",
  scopeId: "",
  eventType: "status_degraded",
  channel: "destination",
  enabled: false,
};
const mountMatrix = (environmentId?: number, disabled = false) =>
  shallowMount(NotificationMatrix, {
    props: { destination, environmentId, disabled },
    global: { renderStubDefaultSlot: true },
  });
beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(api.getNotificationEventTypes).mockResolvedValue({
    data: [{ type: "status_degraded", defaultChannels: ["in_app", "email", "destination"] }],
  } as any);
  vi.mocked(api.getDestinationNotificationPreferences).mockResolvedValue({ data: [] } as any);
  vi.mocked(api.getNotificationPreferences).mockResolvedValue({ data: [] } as any);
  vi.mocked(api.setDestinationNotificationPreference).mockResolvedValue({} as any);
  vi.mocked(api.deleteDestinationNotificationPreference).mockResolvedValue({} as any);
});
it("uses the existing event registry and saves shared defaults separately from personal settings", async () => {
  const w = mountMatrix();
  await flushPromises();
  expect(w.text()).toContain("Status degraded");
  expect(w.findAll("th")).toHaveLength(2);
  const toggle = w.findComponent({ name: "Switch" });
  expect(toggle.props("modelValue")).toBe(true);
  vi.mocked(api.getDestinationNotificationPreferences).mockResolvedValue({
    data: [globalOff],
  } as any);
  toggle.vm.$emit("update:modelValue", false);
  await flushPromises();
  expect(api.setDestinationNotificationPreference).toHaveBeenCalledWith({
    path: { orgId: "org", destinationId: "dest-1" },
    body: { ...globalOff },
  });
  expect(w.findComponent({ name: "Switch" }).props("modelValue")).toBe(false);
  expect(api.setNotificationPreference).not.toHaveBeenCalled();
  w.unmount();
});
it("inherits destination defaults and deletes an environment override to restore inheritance", async () => {
  vi.mocked(api.getDestinationNotificationPreferences).mockResolvedValue({
    data: [globalOff],
  } as any);
  const w = mountMatrix(7);
  await flushPromises();
  const buttons = () => w.findAll("button");
  expect(buttons()[0]!.attributes("aria-pressed")).toBe("true");
  vi.mocked(api.getDestinationNotificationPreferences).mockResolvedValue({
    data: [globalOff, { ...globalOff, scopeType: "environment", scopeId: "7", enabled: true }],
  } as any);
  await buttons()[1]!.trigger("click");
  await flushPromises();
  expect(api.setDestinationNotificationPreference).toHaveBeenCalledWith({
    path: { orgId: "org", destinationId: "dest-1" },
    body: { ...globalOff, scopeType: "environment", scopeId: "7", enabled: true },
  });
  expect(buttons()[1]!.attributes("aria-pressed")).toBe("true");
  await buttons()[0]!.trigger("click");
  await flushPromises();
  expect(api.deleteDestinationNotificationPreference).toHaveBeenCalledWith({
    path: { orgId: "org", destinationId: "dest-1" },
    query: {
      scopeType: "environment",
      scopeId: "7",
      eventType: "status_degraded",
      channel: "destination",
    },
  });
  w.unmount();
});
it("disables shared controls for members", async () => {
  const w = mountMatrix(undefined, true);
  await flushPromises();
  const toggle = w.findComponent({ name: "Switch" });
  expect(toggle.props("disabled")).toBe(true);
  toggle.vm.$emit("update:modelValue", false);
  await flushPromises();
  expect(api.setDestinationNotificationPreference).not.toHaveBeenCalled();
  w.unmount();
});
it("hides controls when preferences cannot be loaded and allows retry", async () => {
  vi.mocked(api.getDestinationNotificationPreferences).mockRejectedValue(new Error("offline"));
  const w = mountMatrix();
  await flushPromises();
  expect(w.find('[role="alert"]').exists()).toBe(true);
  expect(w.find("table").exists()).toBe(false);
  vi.mocked(api.getDestinationNotificationPreferences).mockResolvedValue({ data: [] } as any);
  await w.get("button").trigger("click");
  await flushPromises();
  expect(w.find("table").exists()).toBe(true);
  w.unmount();
});
it("keeps the saved state when a preference update fails", async () => {
  vi.mocked(api.setDestinationNotificationPreference).mockResolvedValue({ error: {} } as any);
  const w = mountMatrix();
  await flushPromises();
  w.findComponent({ name: "Switch" }).vm.$emit("update:modelValue", false);
  await flushPromises();
  expect(alerts.error).toHaveBeenCalled();
  expect(w.findComponent({ name: "Switch" }).props("modelValue")).toBe(true);
  w.unmount();
});
it("keeps shared destinations out of the personal channel matrix", async () => {
  await useNotificationPreferences().loadAll();
  const w = shallowMount(NotificationMatrix);
  await flushPromises();
  expect(w.findAll("th")).toHaveLength(3);
  expect(w.text()).toContain("In-app");
  expect(w.text()).toContain("Email");
  expect(api.getDestinationNotificationPreferences).not.toHaveBeenCalled();
  w.unmount();
});
