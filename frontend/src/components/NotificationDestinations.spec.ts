import { beforeEach, expect, it, vi } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import NotificationDestinations from "./NotificationDestinations.vue";
const alerts = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }));
vi.mock("@/composables/useAlert", () => ({ useAlert: () => alerts }));
vi.mock("@/api/generated", () => ({
  listNotificationDestinations: vi.fn(),
  listShopNotificationDestinations: vi.fn(),
  createNotificationDestination: vi.fn(),
  createShopNotificationDestination: vi.fn(),
  updateNotificationDestination: vi.fn(),
  updateShopNotificationDestination: vi.fn(),
  testNotificationDestination: vi.fn(),
  testShopNotificationDestination: vi.fn(),
  deleteNotificationDestination: vi.fn(),
  deleteShopNotificationDestination: vi.fn(),
}));
import * as api from "@/api/generated";
const destination = {
  id: "dest-1",
  name: "Operations",
  provider: "telegram",
  summary: "12345",
  connectionError: false,
};
const status = (canManage = true, destinations = [destination]) =>
  ({ data: { canManage, destinations } }) as any;
const mountCard = (shopId?: number) =>
  shallowMount(NotificationDestinations, {
    props: { organizationId: "org", shopId },
    global: { renderStubDefaultSlot: true },
  });
beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(api.listNotificationDestinations).mockResolvedValue(status());
  vi.mocked(api.listShopNotificationDestinations).mockResolvedValue(status());
});
it("loads multiple provider summaries without secrets", async () => {
  vi.mocked(api.listNotificationDestinations).mockResolvedValue(
    status(true, [
      destination,
      { ...destination, id: "dest-2", provider: "slack", summary: "Incoming webhook" },
    ]),
  );
  const w = mountCard();
  await flushPromises();
  expect(w.findAll("li")).toHaveLength(2);
  expect(w.text()).toContain("Telegram · 12345");
  expect(w.text()).toContain("Slack · Incoming webhook");
  expect((w.vm as any).connection.token).toBe("");
  w.unmount();
});
it("creates a destination and clears credentials", async () => {
  vi.mocked(api.createNotificationDestination).mockResolvedValue({} as any);
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  vm.add();
  vm.name = " Alerts ";
  vm.chooseProvider("telegram");
  vm.connection.token = " token ";
  vm.connection.channel = "123";
  await vm.save();
  expect(api.createNotificationDestination).toHaveBeenCalledWith({
    path: { orgId: "org" },
    body: { name: "Alerts", dsn: "telegram://token@default?channel=123" },
  });
  expect(vm.connection.token).toBe("");
  expect(vm.formOpen).toBe(false);
  expect(alerts.success).toHaveBeenCalled();
  w.unmount();
});
it("edits the name without replacing the stored DSN", async () => {
  vi.mocked(api.updateNotificationDestination).mockResolvedValue({} as any);
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  vm.edit(destination);
  vm.name = "New name";
  await vm.save();
  expect(api.updateNotificationDestination).toHaveBeenCalledWith({
    path: { orgId: "org", destinationId: "dest-1" },
    body: { name: "New name", dsn: "" },
  });
  w.unmount();
});
it("keeps input after a rejected save without displaying upstream secrets", async () => {
  vi.mocked(api.createNotificationDestination).mockResolvedValue({
    error: { message: "secret" },
  } as any);
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  vm.add();
  vm.name = "Alerts";
  vm.chooseProvider("telegram");
  vm.connection.token = "secret";
  vm.connection.channel = "123";
  await vm.save();
  expect(vm.connection.token).toBe("secret");
  expect(vm.formOpen).toBe(true);
  expect(alerts.error).toHaveBeenCalled();
  expect(alerts.error).not.toHaveBeenCalledWith("secret");
  w.unmount();
});
it("recovers from network failures and releases the busy state", async () => {
  vi.mocked(api.createNotificationDestination).mockRejectedValue(new Error("offline"));
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  vm.add();
  vm.name = "Alerts";
  vm.chooseProvider("telegram");
  vm.connection.token = "secret";
  vm.connection.channel = "123";
  await vm.save();
  expect(vm.isBusy).toBe(false);
  expect(vm.formOpen).toBe(true);
  expect(alerts.error).toHaveBeenCalled();
  w.unmount();
});
it("tests only the selected destination", async () => {
  vi.mocked(api.testNotificationDestination).mockResolvedValue({} as any);
  const w = mountCard();
  await flushPromises();
  await (w.vm as any).sendTest("dest-1");
  expect(api.testNotificationDestination).toHaveBeenCalledWith({
    path: { orgId: "org", destinationId: "dest-1" },
  });
  expect(alerts.success).toHaveBeenCalled();
  w.unmount();
});
it("does not report a failed test as successful", async () => {
  vi.mocked(api.testNotificationDestination).mockResolvedValue({ error: {} } as any);
  const w = mountCard();
  await flushPromises();
  await (w.vm as any).sendTest("dest-1");
  expect(alerts.error).toHaveBeenCalled();
  expect(alerts.success).not.toHaveBeenCalled();
  w.unmount();
});
it("routes all shop operations through the shop endpoints", async () => {
  for (const fn of [
    api.createShopNotificationDestination,
    api.updateShopNotificationDestination,
    api.testShopNotificationDestination,
    api.deleteShopNotificationDestination,
  ])
    vi.mocked(fn).mockResolvedValue({} as any);
  const w = mountCard(42);
  await flushPromises();
  const vm = w.vm as any;
  vm.add();
  vm.name = "Alerts";
  vm.chooseProvider("telegram");
  vm.connection.token = "secret";
  vm.connection.channel = "123";
  await vm.save();
  vm.edit(destination);
  await vm.save();
  await vm.sendTest("dest-1");
  await vm.remove("dest-1");
  expect(api.createShopNotificationDestination).toHaveBeenCalledWith({
    path: { orgId: "org", shopId: 42 },
    body: { name: "Alerts", dsn: "telegram://secret@default?channel=123" },
  });
  for (const fn of [
    api.updateShopNotificationDestination,
    api.testShopNotificationDestination,
    api.deleteShopNotificationDestination,
  ])
    expect(fn).toHaveBeenCalledWith(
      expect.objectContaining({ path: { orgId: "org", shopId: 42, destinationId: "dest-1" } }),
    );
  expect(api.listNotificationDestinations).not.toHaveBeenCalled();
  w.unmount();
});
it("keeps failed loads distinct from an empty destination list", async () => {
  vi.mocked(api.listNotificationDestinations).mockRejectedValue(new Error("offline"));
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  expect(vm.loadFailed).toBe(true);
  expect(vm.isLoading).toBe(false);
  vi.mocked(api.listNotificationDestinations).mockResolvedValue(status());
  await vm.load();
  expect(vm.loadFailed).toBe(false);
  w.unmount();
});
it("respects read-only membership and clears canceled secrets", async () => {
  vi.mocked(api.listNotificationDestinations).mockResolvedValue(status(false));
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  expect(vm.canManage).toBe(false);
  expect(w.findAll("button-stub")).toHaveLength(1);
  await w.find("button-stub").trigger("click");
  expect(w.findComponent({ name: "NotificationMatrix" }).props("disabled")).toBe(true);
  expect(w.text()).toContain("Only organization owners and admins");
  vm.chooseProvider("telegram");
  vm.connection.token = "secret";
  vm.connection.channel = "123";
  vm.cancel();
  expect(vm.connection.token).toBe("");
  w.unmount();
});

it("renders destination names as text and masks connection credentials", async () => {
  vi.mocked(api.listNotificationDestinations).mockResolvedValue(
    status(true, [{ ...destination, name: "<img src=x onerror=alert(1)>" }]),
  );
  const w = mountCard();
  await flushPromises();
  expect(w.find("img").exists()).toBe(false);
  expect(w.text()).toContain("<img src=x onerror=alert(1)>");
  (w.vm as any).add();
  await flushPromises();
  expect(w.find('input-stub[type="password"]').exists()).toBe(true);
  w.unmount();
});

it("builds a Slack DSN from a normal pasted webhook URL", async () => {
  vi.mocked(api.createNotificationDestination).mockResolvedValue({} as any);
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  vm.add();
  vm.name = "Slack alerts";
  vm.connection.webhookUrl = " https://hooks.slack.com/services/T/B/secret ";
  await vm.save();
  expect(api.createNotificationDestination).toHaveBeenCalledWith({
    path: { orgId: "org" },
    body: {
      name: "Slack alerts",
      dsn: "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2FT%2FB%2Fsecret",
    },
  });
  expect(vm.connection.webhookUrl).toBe("");
  w.unmount();
});
it("shows only fields for the chosen service and clears previous credentials", async () => {
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  vm.add();
  vm.connection.webhookUrl = "secret";
  vm.chooseProvider("gotify");
  await flushPromises();
  expect(vm.connection.webhookUrl).toBe("");
  expect(w.text()).toContain("Server URL");
  expect(w.text()).toContain("Application token");
  expect(w.text()).not.toContain("Webhook URL");
  expect(w.text()).not.toContain("Connection DSN");
  w.unmount();
});
it("validates a webhook before sending it to the API", async () => {
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  vm.add();
  vm.name = "Alerts";
  vm.connection.webhookUrl = "https://example.com/secret";
  await vm.save();
  await flushPromises();
  expect(api.createNotificationDestination).not.toHaveBeenCalled();
  expect(w.find('[role="alert"]').text()).toContain("Slack incoming webhook URL");
  expect(w.find('[role="alert"]').text()).not.toContain("secret");
  w.unmount();
});
it("requires an explicit connection replacement when editing", async () => {
  vi.mocked(api.updateNotificationDestination).mockResolvedValue({} as any);
  const w = mountCard();
  await flushPromises();
  const vm = w.vm as any;
  vm.edit(destination);
  await flushPromises();
  expect(w.text()).toContain("Your saved connection will be kept");
  expect(w.find('input-stub[type="password"]').exists()).toBe(false);
  vm.replaceConnection = true;
  vm.connection.token = "new:token";
  vm.connection.channel = "-123";
  await vm.save();
  expect(api.updateNotificationDestination).toHaveBeenCalledWith({
    path: { orgId: "org", destinationId: "dest-1" },
    body: { name: "Operations", dsn: "telegram://new%3Atoken@default?channel=-123" },
  });
  w.unmount();
});

it("lets admins replace an unreadable connection", async () => {
  vi.mocked(api.listNotificationDestinations).mockResolvedValue(
    status(true, [{ ...destination, provider: "", summary: "", connectionError: true }]),
  );
  const w = mountCard();
  await flushPromises();
  expect(w.text()).toContain("Connection unavailable");
  (w.vm as any).edit({ ...destination, connectionError: true });
  await flushPromises();
  expect(w.find('input-stub[type="password"]').exists()).toBe(true);
  w.unmount();
});
