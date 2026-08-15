import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ref } from "vue";
import DetailUptime from "./DetailUptime.vue";

vi.mock("chart.js", () => {
  class Chart {
    static register = vi.fn();
    destroy = vi.fn();
  }
  return {
    Chart,
    registerables: [],
  };
});

vi.mock("chartjs-adapter-date-fns", () => ({}));

vi.mock("@/composables/useEnvironmentDetail", () => ({
  useEnvironmentDetail: () => ({
    environment: ref({
      id: 1,
      name: "Production",
      url: "https://shop.example.com",
    }),
  }),
}));

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getEnvironmentUptime: vi.fn(),
  updateUptimeSettings: vi.fn(),
}));

import { getEnvironmentUptime, updateUptimeSettings } from "@/api/generated";

function uptimeResponse(overrides: Record<string, unknown> = {}) {
  return {
    settings: {
      enabled: true,
      url: null,
      intervalSeconds: 60,
      expectedStatus: 0,
      contentMatch: null,
      failureThreshold: 3,
      recoveryThreshold: 2,
      status: "up",
      lastCheckedAt: "2026-08-15T12:00:00Z",
      lastStatusCode: 200,
      lastLatencyMs: 123,
      lastError: null,
    },
    openIncident: null,
    availability: 0.9994,
    days: [],
    incidents: [],
    latency: [],
    ...overrides,
  };
}

function mountComponent() {
  return mount(DetailUptime, {
    global: {
      stubs: {
        // Interactive reka-ui primitives are not the subject of these tests.
        Select: { template: "<div><slot /></div>" },
        SelectTrigger: { template: "<div><slot /></div>" },
        SelectValue: { template: "<div />" },
        SelectContent: { template: "<div><slot /></div>" },
        SelectItem: { template: "<div><slot /></div>" },
        Switch: { template: "<input type='checkbox' />" },
        "icon-fa6-solid:rotate": { template: "<span />" },
      },
    },
  });
}

describe("DetailUptime", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getEnvironmentUptime).mockResolvedValue({
      data: uptimeResponse(),
      error: undefined,
      response: new Response(),
    } as never);
    vi.mocked(updateUptimeSettings).mockResolvedValue({
      data: undefined,
      error: undefined,
      response: new Response(),
    } as never);
  });

  it("renders status and availability when monitoring is enabled", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Uptime monitoring");
    expect(wrapper.text()).toContain("Up");
    expect(wrapper.text()).toContain("99.940%");
    expect(wrapper.text()).toContain("123ms");

    expect(getEnvironmentUptime).toHaveBeenCalledWith(
      expect.objectContaining({
        path: { environmentId: 1 },
        query: { range: "24h" },
      }),
    );
  });

  it("shows the enable hint when monitoring is disabled", async () => {
    vi.mocked(getEnvironmentUptime).mockResolvedValue({
      data: uptimeResponse({
        settings: { ...uptimeResponse().settings, enabled: false, status: "unknown" },
        availability: null,
      }),
      error: undefined,
      response: new Response(),
    } as never);

    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Disabled");
    expect(wrapper.text()).toContain("Enable uptime monitoring");
  });

  it("shows the open incident banner and incident list", async () => {
    const incident = {
      id: 7,
      startedAt: "2026-08-15T11:30:00Z",
      resolvedAt: null,
      durationSeconds: 1800,
      statusCode: 503,
      latencyMs: null,
      error: "status 503",
    };
    vi.mocked(getEnvironmentUptime).mockResolvedValue({
      data: uptimeResponse({
        settings: { ...uptimeResponse().settings, status: "down" },
        openIncident: incident,
        incidents: [incident],
      }),
      error: undefined,
      response: new Response(),
    } as never);

    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Down");
    expect(wrapper.text()).toContain("Ongoing");
    expect(wrapper.text()).toContain("HTTP 503");
    expect(wrapper.text()).toContain("30m 0s");
  });

  it("renders the day strip for multi-day ranges", async () => {
    vi.mocked(getEnvironmentUptime).mockResolvedValue({
      data: uptimeResponse({
        days: [
          { day: "2026-08-14", availability: 1, downSeconds: 0 },
          { day: "2026-08-15", availability: 0.95, downSeconds: 3600 },
        ],
      }),
      error: undefined,
      response: new Response(),
    } as never);

    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Daily availability");
    expect(wrapper.text()).toContain("2026-08-14");
    expect(wrapper.text()).toContain("2026-08-15");
  });

  it("saves settings and reloads", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    const urlInput = wrapper.find("#uptime-url");
    await urlInput.setValue("https://status.example.com/health");

    await wrapper.find("button").trigger("click");
    await flushPromises();

    expect(updateUptimeSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        path: { environmentId: 1 },
        body: expect.objectContaining({
          enabled: true,
          url: "https://status.example.com/health",
          intervalSeconds: 60,
          failureThreshold: 3,
          recoveryThreshold: 2,
        }),
      }),
    );
    // load() is called once on mount and once after saving
    expect(getEnvironmentUptime).toHaveBeenCalledTimes(2);
  });

  it("flips the enable switch and sends the new value", async () => {
    // Regression: the switch must bind via v-model (modelValue), not a
    // checked prop the reka-ui SwitchRoot does not have. Render the real
    // component to catch binding drift.
    vi.mocked(getEnvironmentUptime).mockResolvedValue({
      data: uptimeResponse({ settings: { ...uptimeResponse().settings, enabled: false } }),
      error: undefined,
      response: new Response(),
    } as never);

    const wrapper = mount(DetailUptime, {
      global: {
        stubs: {
          Select: { template: "<div><slot /></div>" },
          SelectTrigger: { template: "<div><slot /></div>" },
          SelectValue: { template: "<div />" },
          SelectContent: { template: "<div><slot /></div>" },
          SelectItem: { template: "<div><slot /></div>" },
          "icon-fa6-solid:rotate": { template: "<span />" },
        },
      },
    });
    await flushPromises();

    const toggle = wrapper.find("[role='switch']");
    expect(toggle.exists()).toBe(true);
    expect(toggle.attributes("data-state")).toBe("unchecked");

    await toggle.trigger("click");
    await flushPromises();
    expect(wrapper.find("[role='switch']").attributes("data-state")).toBe("checked");

    const saveButton = wrapper.findAll("button").find((b) => b.text().includes("Save settings"));
    expect(saveButton).toBeDefined();
    await saveButton!.trigger("click");
    await flushPromises();

    expect(updateUptimeSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        body: expect.objectContaining({ enabled: true }),
      }),
    );
  });
});
