import { beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, shallowMount } from "@vue/test-utils";
import { ref } from "vue";
import DetailChecks from "./DetailChecks.vue";
import DetailScheduledTasks from "./DetailScheduledTasks.vue";
import DetailDeployments from "./DetailDeployments.vue";

const mocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn(), info: vi.fn() }));
const environment = ref<any>(null);
vi.mock("@/composables/useEnvironmentDetail", () => ({
  useEnvironmentDetail: () => ({ environment }),
}));
vi.mock("@/composables/useAlert", () => ({ useAlert: () => mocks }));
vi.mock("vue-router", async (importOriginal) => ({
  ...(await importOriginal<typeof import("vue-router")>()),
  useRoute: () => ({ params: {} }),
}));
vi.mock("@/api/generated", () => ({
  updateEnvironment: vi.fn(),
  rescheduleTask: vi.fn(),
  deleteDeployment: vi.fn(),
  getDeployments: vi.fn(),
}));
import {
  updateEnvironment,
  rescheduleTask,
  deleteDeployment,
  getDeployments,
} from "@/api/generated";

describe("environment mutation failures", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    environment.value = { id: 1, shopId: 2, checks: [], ignores: [], scheduledTasks: [] };
    vi.mocked(getDeployments).mockResolvedValue({ data: [{ id: 3, returnCode: 0 }] } as any);
  });
  it.each(["ignoreCheck", "removeIgnore"])(
    "%s preserves ignore state on resolved error",
    async (action) => {
      environment.value.ignores = ["existing"];
      vi.mocked(updateEnvironment).mockResolvedValue({
        error: { message: "Access denied" },
      } as any);
      const wrapper = shallowMount(DetailChecks);
      await (wrapper.vm as any)[action]("existing");
      expect(environment.value.ignores).toEqual(["existing"]);
      expect(mocks.info).not.toHaveBeenCalled();
      expect(mocks.error).toHaveBeenCalledWith("Access denied");
      wrapper.unmount();
    },
  );
  it("does not announce a rejected task reschedule", async () => {
    vi.mocked(rescheduleTask).mockResolvedValue({ error: "Service unavailable" } as any);
    const wrapper = shallowMount(DetailScheduledTasks);
    await (wrapper.vm as any).onReScheduleTask("task");
    expect(mocks.success).not.toHaveBeenCalled();
    expect(mocks.error).toHaveBeenCalledWith("Service unavailable");
    wrapper.unmount();
  });
  it("keeps deletion dialog and deployments on failure and permits retry", async () => {
    vi.mocked(deleteDeployment)
      .mockResolvedValueOnce({ error: {} } as any)
      .mockResolvedValueOnce({} as any);
    const wrapper = shallowMount(DetailDeployments);
    await flushPromises();
    const vm = wrapper.vm as any;
    vm.confirmDeleteDeployment({ id: 3 });
    await vm.deleteDeployment();
    expect(vm.showDeleteDeploymentDialog).toBe(true);
    expect(vm.deploymentToDelete.id).toBe(3);
    expect(vm.deployments).toHaveLength(1);
    expect(vm.isDeletingDeployment).toBe(false);
    expect(getDeployments).toHaveBeenCalledTimes(1);
    expect(mocks.error).toHaveBeenCalledWith(expect.stringContaining("Please try again"));
    await vm.deleteDeployment();
    expect(vm.showDeleteDeploymentDialog).toBe(false);
    expect(getDeployments).toHaveBeenCalledTimes(2);
    wrapper.unmount();
  });
});
