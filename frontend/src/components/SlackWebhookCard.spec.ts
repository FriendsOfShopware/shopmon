import { beforeEach, expect, it, vi } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import SlackWebhookCard from "./SlackWebhookCard.vue";
const mocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }));
vi.mock("@/composables/useAlert", () => ({ useAlert: () => mocks }));
vi.mock("@/api/generated", () => ({
  getSlackWebhook: vi.fn(),
  updateSlackWebhook: vi.fn(),
  testSlackWebhook: vi.fn(),
  deleteSlackWebhook: vi.fn(),
  getShopSlackWebhook: vi.fn(),
  updateShopSlackWebhook: vi.fn(),
  testShopSlackWebhook: vi.fn(),
  deleteShopSlackWebhook: vi.fn(),
}));
import {
  deleteSlackWebhook,
  getShopSlackWebhook,
  getSlackWebhook,
  testSlackWebhook,
  updateShopSlackWebhook,
  updateSlackWebhook,
} from "@/api/generated";
const mountCard = (shopId?: number) =>
  shallowMount(SlackWebhookCard, { props: { organizationId: "org", shopId } });
const status = (configured: boolean, canManage = true) =>
  ({ data: { configured, canManage } }) as any;
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getSlackWebhook).mockResolvedValue(status(false));
  vi.mocked(getShopSlackWebhook).mockResolvedValue(status(false));
});
it("saves the webhook and clears the input", async () => {
  vi.mocked(updateSlackWebhook).mockResolvedValue({} as any);
  const wrapper = mountCard();
  await flushPromises();
  const vm = wrapper.vm as any;
  vm.webhookUrl = "  https://hooks.slack.com/services/a  ";
  await vm.save();
  expect(updateSlackWebhook).toHaveBeenCalledWith({
    path: { orgId: "org" },
    body: { webhookUrl: "https://hooks.slack.com/services/a" },
  });
  expect(vm.configured).toBe(true);
  expect(vm.webhookUrl).toBe("");
  expect(mocks.success).toHaveBeenCalled();
  wrapper.unmount();
});
it("keeps the input and shows the API message when saving fails", async () => {
  vi.mocked(updateSlackWebhook).mockResolvedValue({
    error: { message: "webhook URL must start with https://hooks.slack.com/" },
  } as any);
  const wrapper = mountCard();
  await flushPromises();
  const vm = wrapper.vm as any;
  vm.webhookUrl = "https://example.com/hook";
  await vm.save();
  expect(vm.configured).toBe(false);
  expect(vm.webhookUrl).toBe("https://example.com/hook");
  expect(mocks.error).toHaveBeenCalledWith("webhook URL must start with https://hooks.slack.com/");
  wrapper.unmount();
});
it("reports a rejected test message", async () => {
  vi.mocked(getSlackWebhook).mockResolvedValue(status(true));
  vi.mocked(testSlackWebhook).mockResolvedValue({
    error: { message: "Slack rejected the test message; check the webhook URL" },
  } as any);
  const wrapper = mountCard();
  await flushPromises();
  await (wrapper.vm as any).sendTest();
  expect(mocks.error).toHaveBeenCalledWith(
    "Slack rejected the test message; check the webhook URL",
  );
  expect(mocks.success).not.toHaveBeenCalled();
  wrapper.unmount();
});
it("disconnects", async () => {
  vi.mocked(getSlackWebhook).mockResolvedValue(status(true));
  vi.mocked(deleteSlackWebhook).mockResolvedValue({} as any);
  const wrapper = mountCard();
  await flushPromises();
  const vm = wrapper.vm as any;
  expect(vm.configured).toBe(true);
  await vm.remove();
  expect(vm.configured).toBe(false);
  wrapper.unmount();
});
it("hides the form from members who cannot manage", async () => {
  vi.mocked(getSlackWebhook).mockResolvedValue(status(false, false));
  const wrapper = mountCard();
  await flushPromises();
  expect(wrapper.find("form").exists()).toBe(false);
  wrapper.unmount();
});
it("manages the shop webhook when a shop is given", async () => {
  vi.mocked(updateShopSlackWebhook).mockResolvedValue({} as any);
  const wrapper = mountCard(5);
  await flushPromises();
  const vm = wrapper.vm as any;
  expect(getShopSlackWebhook).toHaveBeenCalledWith({ path: { orgId: "org", shopId: 5 } });
  expect(getSlackWebhook).not.toHaveBeenCalled();
  vm.webhookUrl = "https://hooks.slack.com/services/shop";
  await vm.save();
  expect(updateShopSlackWebhook).toHaveBeenCalledWith({
    path: { orgId: "org", shopId: 5 },
    body: { webhookUrl: "https://hooks.slack.com/services/shop" },
  });
  expect(updateSlackWebhook).not.toHaveBeenCalled();
  wrapper.unmount();
});
