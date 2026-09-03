import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ref } from "vue";
import type { ChartConfiguration } from "chart.js";

// Chart.js needs a real canvas, so record the configs it would render instead.
const chartConfigs: ChartConfiguration[] = [];

vi.mock("chart.js", () => {
  class ChartStub {
    static register = vi.fn();
    config: ChartConfiguration;
    constructor(_ctx: unknown, config: ChartConfiguration) {
      this.config = config;
      chartConfigs.push(config);
    }
    destroy() {}
  }
  return { Chart: ChartStub, registerables: [] };
});
vi.mock("chartjs-plugin-annotation", () => ({ default: {} }));
vi.mock("chartjs-adapter-date-fns", () => ({}));

const DAY = 24 * 60 * 60 * 1000;

function daysAgo(days: number): string {
  return new Date(Date.now() - days * DAY).toISOString();
}

// Two runs inside the default 30 day window, one far outside it.
const recentRun = {
  createdAt: daysAgo(1),
  ttfb: 100,
  fullyLoaded: 1000,
  largestContentfulPaint: 900,
  firstContentfulPaint: 800,
  cumulativeLayoutShift: 0.01,
  transferSize: 1024,
};
const secondRecentRun = {
  createdAt: daysAgo(5),
  ttfb: 200,
  fullyLoaded: 2000,
  largestContentfulPaint: 1900,
  firstContentfulPaint: 1800,
  cumulativeLayoutShift: 0.02,
  transferSize: 2048,
};
const oldRun = {
  createdAt: daysAgo(120),
  ttfb: 300,
  fullyLoaded: 3000,
  largestContentfulPaint: 2900,
  firstContentfulPaint: 2800,
  cumulativeLayoutShift: 0.03,
  transferSize: 3072,
};

const environment = ref<Record<string, unknown> | null>(null);

vi.mock("@/composables/useEnvironmentDetail", () => ({
  useEnvironmentDetail: () => ({ environment }),
}));

import DetailSitespeed from "./DetailSitespeed.vue";

// The component only builds its charts once the canvas reports a real size,
// which jsdom never does on its own.
function stubCanvasLayout() {
  for (const prop of ["clientWidth", "clientHeight"]) {
    Object.defineProperty(HTMLCanvasElement.prototype, prop, {
      configurable: true,
      get: () => 600,
    });
  }
  HTMLCanvasElement.prototype.getContext = vi.fn(
    () => ({}) as unknown as CanvasRenderingContext2D,
  ) as unknown as HTMLCanvasElement["getContext"];
}

/** x values (timestamps) of the first dataset of the most recent chart render. */
function lastChartTimestamps(): number[] {
  const config = chartConfigs[chartConfigs.length - 1];
  const points = config.data.datasets[0].data as Array<{ x: number }>;
  return points.map((point) => point.x);
}

async function settleCharts() {
  await flushPromises();
  // Two nested requestAnimationFrame hops before the charts are created.
  await new Promise((resolve) => setTimeout(resolve, 50));
}

function mountComponent() {
  return mount(DetailSitespeed);
}

describe("DetailSitespeed", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    chartConfigs.length = 0;
    stubCanvasLayout();
    environment.value = {
      id: 1,
      sitespeedEnabled: true,
      sitespeedDetailUrl: "https://sitespeed.example/result/index.html",
      sitespeeds: [recentRun, secondRecentRun, oldRun],
    };
  });

  it("renders a visibility toggle for every history row", async () => {
    const wrapper = mountComponent();
    await settleCharts();

    const toggles = wrapper.findAll('button[aria-label="Hide this run from the charts"]');
    expect(toggles).toHaveLength(3);
  });

  it("defaults the timespan to the last 30 days", async () => {
    const wrapper = mountComponent();
    await settleCharts();

    expect(wrapper.text()).toContain("Last 30 days");
    // The 120 day old run is outside the default window.
    expect(lastChartTimestamps()).toEqual([
      new Date(secondRecentRun.createdAt).getTime(),
      new Date(recentRun.createdAt).getTime(),
    ]);
  });

  it("drops a hidden run from the charts and keeps it in the table", async () => {
    const wrapper = mountComponent();
    await settleCharts();

    await wrapper.findAll('button[aria-label="Hide this run from the charts"]')[0].trigger("click");
    await settleCharts();

    expect(lastChartTimestamps()).toEqual([new Date(secondRecentRun.createdAt).getTime()]);
    expect(wrapper.findAll("tbody tr")).toHaveLength(3);
    expect(wrapper.findAll("tbody tr")[0].classes()).toContain("opacity-50");
  });

  it("flips the toggle to a show action once a run is hidden", async () => {
    const wrapper = mountComponent();
    await settleCharts();

    await wrapper.findAll('button[aria-label="Hide this run from the charts"]')[0].trigger("click");
    await settleCharts();

    const toggle = wrapper.findAll("tbody tr")[0].find("button");
    expect(toggle.attributes("aria-label")).toBe("Show this run in the charts");
    expect(toggle.attributes("aria-pressed")).toBe("true");
  });

  it("restores every hidden run via 'Show all runs'", async () => {
    const wrapper = mountComponent();
    await settleCharts();

    await wrapper.findAll('button[aria-label="Hide this run from the charts"]')[0].trigger("click");
    await settleCharts();

    const showAll = wrapper.findAll("button").find((button) => button.text() === "Show all runs");
    expect(showAll).toBeTruthy();

    await showAll!.trigger("click");
    await settleCharts();

    expect(lastChartTimestamps()).toEqual([
      new Date(secondRecentRun.createdAt).getTime(),
      new Date(recentRun.createdAt).getTime(),
    ]);
    expect(wrapper.findAll("button").find((button) => button.text() === "Show all runs")).toBe(
      undefined,
    );
  });

  it("explains an empty result instead of rendering blank charts", async () => {
    environment.value = { ...environment.value, sitespeeds: [oldRun] };
    const wrapper = mountComponent();
    await settleCharts();

    // The single run is 120 days old, so the default 30 day window is empty.
    expect(wrapper.find("canvas").exists()).toBe(false);
    expect(wrapper.text()).toContain("No runs in the selected timespan");

    (wrapper.vm as unknown as { timespan: string }).timespan = "all";
    await settleCharts();

    expect(wrapper.find("canvas").exists()).toBe(true);
    expect(wrapper.text()).not.toContain("No runs in the selected timespan");
  });

  it("includes older runs when the timespan is widened", async () => {
    const wrapper = mountComponent();
    await settleCharts();

    // The Select renders a listbox only when opened, so drive the model directly.
    (wrapper.vm as unknown as { timespan: string }).timespan = "all";
    await settleCharts();

    expect(lastChartTimestamps()).toEqual([
      new Date(oldRun.createdAt).getTime(),
      new Date(secondRecentRun.createdAt).getTime(),
      new Date(recentRun.createdAt).getTime(),
    ]);
  });
});
