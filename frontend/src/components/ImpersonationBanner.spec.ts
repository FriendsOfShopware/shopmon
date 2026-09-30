import { mount } from "@vue/test-utils";
import { ref } from "vue";
import { describe, expect, it, vi } from "vitest";
import ImpersonationBanner from "./ImpersonationBanner.vue";

const maliciousEmail = '<img src=x onerror="globalThis.exfiltrated=true">';

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: ref({
      user: { email: maliciousEmail },
      session: { impersonatedBy: "admin-id" },
    }),
  }),
}));

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({ error: vi.fn() }),
}));

vi.mock("@/api/generated", () => ({
  adminStopImpersonating: vi.fn(),
}));

describe("ImpersonationBanner", () => {
  it("renders the impersonated user's email as text", () => {
    const wrapper = mount(ImpersonationBanner);

    expect(wrapper.text()).toContain(maliciousEmail);
    expect(wrapper.find("img").exists()).toBe(false);
  });
});
