import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises, enableAutoUnmount } from "@vue/test-utils";
import { reactive, ref } from "vue";

enableAutoUnmount(afterEach);

const getEnvironmentAccounts = vi.fn();
const acknowledgeEnvironmentAccountEvent = vi.fn();

vi.mock("@/api/generated", () => ({
  getEnvironmentAccounts: (...args: unknown[]) => getEnvironmentAccounts(...args),
  acknowledgeEnvironmentAccountEvent: (...args: unknown[]) =>
    acknowledgeEnvironmentAccountEvent(...args),
}));

// Reactive, so changing it drives the component's route watcher the way
// navigating between environments does.
const routeParams = reactive({ environmentId: "7" });
vi.mock("vue-router", () => ({
  useRoute: () => ({ params: routeParams }),
}));

const environment = ref<Record<string, unknown> | null>({ id: 7, openAccountFindings: 1 });
const loadEnvironment = vi.fn();
vi.mock("@/composables/useEnvironmentDetail", () => ({
  useEnvironmentDetail: () => ({ environment, loadEnvironment }),
}));

const alerts = { info: vi.fn(), error: vi.fn() };
vi.mock("@/composables/useAlert", () => ({ useAlert: () => alerts }));

import DetailAccounts from "./DetailAccounts.vue";

const adminFinding = {
  id: 42,
  level: "red",
  kind: "user",
  change: "new_admin_user",
  label: "b***@e***.de",
  messageKey: "check.account.newAdminUser",
  params: { label: "b***@e***.de" },
  adminUrl: "https://shop.example.com/admin#/sw/users/permissions/user/detail/abc",
  detectedAt: "2026-10-29T07:00:00Z",
  checkId: "account.user.abc.new_admin_user",
  revision: 3,
  account: {
    active: true,
    admin: true,
    roleCount: 0,
    accessKeyCount: 2,
    privilegeCount: 0,
    managedByApp: false,
    passwordRecoveryPending: false,
    createdAt: "2026-10-29T06:00:00Z",
    firstSeenAt: "2026-10-29T07:00:00Z",
  },
};

function response(overrides: Record<string, unknown> = {}) {
  return {
    data: {
      enabled: true,
      supported: true,
      baselineAt: "2026-10-01T00:00:00Z",
      lastSyncedAt: "2026-10-29T07:00:00Z",
      accountCount: 4,
      openFindings: [],
      events: [],
      postureHints: [],
      ...overrides,
    },
  };
}

async function mountComponent() {
  const wrapper = mount(DetailAccounts, { attachTo: document.body });
  await flushPromises();
  return wrapper;
}

/** The dialog renders through a Teleport, so it lives outside the wrapper. */
function bodyButton(label: string): HTMLButtonElement | undefined {
  return [...document.body.querySelectorAll("button")].find((b) => b.textContent?.trim() === label);
}

describe("DetailAccounts", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    environment.value = { id: 7, openAccountFindings: 1 };
    document.body.innerHTML = "";
    routeParams.environmentId = "7";
    getEnvironmentAccounts.mockResolvedValue(response());
    acknowledgeEnvironmentAccountEvent.mockResolvedValue({
      data: undefined,
      error: undefined,
      response: new Response(null, { status: 204 }),
    });
  });

  it("renders a finding with its masked label and a deep link into Shopware", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ openFindings: [adminFinding] }));

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("New administrator account b***@e***.de was created");
    const link = wrapper.find(`a[href="${adminFinding.adminUrl}"]`);
    expect(link.exists()).toBe(true);
    expect(link.attributes("rel")).toContain("noopener");
  });

  it("offers no deep link when the shop supplied no usable admin path", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({ openFindings: [{ ...adminFinding, adminUrl: undefined }] }),
    );

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("New administrator account");
    expect(wrapper.find('a[target="_blank"]').exists()).toBe(false);
  });

  it("acknowledges a finding and updates the tab badge", async () => {
    getEnvironmentAccounts
      .mockResolvedValueOnce(response({ openFindings: [adminFinding] }))
      .mockResolvedValueOnce(response({ openFindings: [] }));

    const wrapper = await mountComponent();
    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Looks right")
      ?.trigger("click");
    await flushPromises();

    const confirm = bodyButton("Confirm");
    expect(confirm).toBeTruthy();
    confirm?.click();
    await flushPromises();

    expect(acknowledgeEnvironmentAccountEvent).toHaveBeenCalledWith({
      path: { environmentId: 7, eventId: 42 },
      body: { note: undefined, revision: 3 },
    });
    expect(environment.value?.openAccountFindings).toBe(0);
  });

  it("explains that nothing is reported before the first scrape", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ baselineAt: undefined }));

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("Not initialised yet");
  });

  it("points at a plugin update when the endpoint is not served", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ supported: false }));

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("FroshTools is too old");
  });

  it("surfaces the reason the last read failed", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({ lastError: "shopware api error (status 403): missing frosh_tools:read" }),
    );

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("frosh_tools:read");
  });

  it("sends the reader to the environment settings when monitoring is off", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ enabled: false }));

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("Account monitoring is switched off");
    // The switch itself lives in the environment settings, next to sitespeed.
    expect(wrapper.text()).toContain("Open settings");
  });

  it("renders posture hints as advice rather than findings", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({
        postureHints: [
          {
            hint: "unusedAdminIntegration",
            kind: "integration",
            label: "Erp Connector",
            detail: "never used",
          },
        ],
      }),
    );

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("Administrator integration has not been used in a long time");
    expect(wrapper.text()).toContain("Erp Connector");
    // Hints must not be presented as something to confirm.
    expect(wrapper.findAll("button").some((b) => b.text() === "Looks right")).toBe(false);
  });

  it("shows informational history entries without offering to confirm them", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({
        events: [
          {
            id: 9,
            level: "info",
            kind: "user",
            change: "admin_revoked",
            label: "a***@e***.de",
            messageKey: "check.account.adminRevoked",
            params: { label: "a***@e***.de" },
            detectedAt: "2026-10-28T07:00:00Z",
          },
        ],
      }),
    );

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("a***@e***.de lost administrator rights");
    expect(wrapper.findAll("button").some((b) => b.text() === "Looks right")).toBe(false);
  });

  it("shows the summary tiles the other tabs use", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({ openFindings: [adminFinding], accountCount: 9 }),
    );

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("Critical");
    expect(wrapper.text()).toContain("Monitored accounts");
    expect(wrapper.text()).toContain("9");
  });

  // Without these a reviewer only sees "a new admin was created" and has to
  // follow the deep link to learn anything about the account.
  it("lists the account's current state next to the finding", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ openFindings: [adminFinding] }));

    const wrapper = await mountComponent();
    const text = wrapper.text();

    expect(text).toContain("b***@e***.de");
    expect(text).toContain("Administrator");
    expect(text).toContain("API keys");
    expect(text).toContain("Created in the shop");
  });

  it("says so when the account behind a finding is already gone", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({ openFindings: [{ ...adminFinding, account: undefined }] }),
    );

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("no longer exists");
    expect(wrapper.text()).toContain("b***@e***.de");
  });

  it("renders hint facts in the viewer's language rather than a server string", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({
        postureHints: [
          { hint: "unusedAdminIntegration", kind: "integration", label: "Erp Connector" },
          {
            hint: "userWithApiKeys",
            kind: "user",
            label: "a***@e.de",
            accessKeyCount: 3,
          },
        ],
      }),
    );

    const wrapper = await mountComponent();
    const text = wrapper.text();

    expect(text).toContain("Last used");
    expect(text).toContain("never");
    expect(text).toContain("API keys");
    expect(text).toContain("3");
  });

  it("filters the history down to findings or informational entries", async () => {
    const finding = { ...adminFinding, id: 1 };
    const info = {
      id: 2,
      level: "info",
      kind: "user",
      change: "admin_revoked",
      label: "a***@e***.de",
      messageKey: "check.account.adminRevoked",
      params: { label: "a***@e***.de" },
      detectedAt: "2026-10-28T07:00:00Z",
    };
    getEnvironmentAccounts.mockResolvedValue(response({ events: [finding, info] }));

    const wrapper = await mountComponent();
    expect(wrapper.text()).toContain("lost administrator rights");

    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Needs review")
      ?.trigger("click");
    expect(wrapper.text()).not.toContain("lost administrator rights");

    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Informational")
      ?.trigger("click");
    expect(wrapper.text()).toContain("lost administrator rights");
    expect(wrapper.text()).not.toContain("New administrator account");
  });

  // Every sibling tab (checks, tasks, queue, status history) renders separate
  // bordered rows rather than one divided block. There is no shared component
  // for it, so the convention is only held by matching it.
  it("renders list rows the way the sibling tabs do", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({
        openFindings: [adminFinding],
        postureHints: [
          { hint: "userWithApiKeys", kind: "user", label: "a***@e.de", accessKeyCount: 1 },
        ],
        events: [
          {
            id: 2,
            level: "info",
            kind: "user",
            change: "admin_revoked",
            label: "a***@e***.de",
            messageKey: "check.account.adminRevoked",
            params: { label: "a***@e***.de" },
            detectedAt: "2026-10-28T07:00:00Z",
          },
        ],
      }),
    );

    const wrapper = await mountComponent();

    expect(wrapper.html()).not.toContain("divide-y");
    // One row per finding, hint and history entry, each with its own border.
    expect(wrapper.findAll(".rounded-xl.border").length).toBeGreaterThanOrEqual(3);
  });

  // The generated client reports failures through `error` rather than throwing,
  // so a catch block alone would let a failed request render as an empty tab.
  it("shows a failure state when the request errors", async () => {
    getEnvironmentAccounts.mockResolvedValue({
      data: undefined,
      error: { message: "boom" },
    });

    const wrapper = await mountComponent();

    expect(wrapper.text()).toContain("The accounts could not be loaded");
    expect(wrapper.text()).toContain("Retry");
    expect(alerts.error).toHaveBeenCalled();
  });

  // The layout reuses this page across environments; a slower request for the
  // previous one must not replace the current findings or the badge count.
  it("discards a response that arrived after the environment changed", async () => {
    let release: (v: unknown) => void = () => {};
    getEnvironmentAccounts.mockImplementationOnce(
      () => new Promise((resolve) => (release = resolve)),
    );

    const wrapper = mount(DetailAccounts, { attachTo: document.body });
    // The reader moves on before the first response lands.
    routeParams.environmentId = "8";
    release(response({ openFindings: [adminFinding], accountCount: 99 }));
    await flushPromises();

    expect(wrapper.text()).not.toContain("New administrator account");
    // The switch triggers a load for the new environment; what must not happen
    // is the stale response's count landing in the shared badge.
    expect(environment.value?.openAccountFindings).not.toBe(99);
  });

  // Confirming queues a scrape that recomputes the status, so the shared
  // environment has to be reloaded or the header keeps its stale colour.
  it("reloads the shared environment after confirming", async () => {
    getEnvironmentAccounts
      .mockResolvedValueOnce(response({ openFindings: [adminFinding] }))
      .mockResolvedValueOnce(response({ openFindings: [] }));

    const wrapper = await mountComponent();
    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Looks right")
      ?.trigger("click");
    await flushPromises();
    bodyButton("Confirm")?.click();
    await flushPromises();

    expect(loadEnvironment).toHaveBeenCalledWith(true);
  });

  // "Needs review" has to mean the same thing as the list above it.
  it("keeps acknowledged entries out of the needs-review filter", async () => {
    const acknowledged = {
      ...adminFinding,
      id: 3,
      acknowledgedAt: "2026-10-30T07:00:00Z",
      acknowledgedBy: "Owner",
    };
    getEnvironmentAccounts.mockResolvedValue(response({ events: [acknowledged] }));

    const wrapper = await mountComponent();
    expect(wrapper.text()).toContain("New administrator account");

    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Needs review")
      ?.trigger("click");

    expect(wrapper.text()).not.toContain("New administrator account");
  });

  // Two loads for the same environment also race: this view reloads after a
  // confirmation while the route watcher may fire as well.
  it("discards a superseded response for the same environment", async () => {
    let releaseFirst: (v: unknown) => void = () => {};
    getEnvironmentAccounts
      .mockImplementationOnce(() => new Promise((resolve) => (releaseFirst = resolve)))
      .mockResolvedValueOnce(response({ openFindings: [], accountCount: 2 }));

    const wrapper = mount(DetailAccounts, { attachTo: document.body });
    // A second load starts and finishes before the first one answers.
    const vm = wrapper.vm as unknown as { load: () => Promise<void> };
    expect(typeof vm.load).toBe("function");
    await vm.load();
    releaseFirst(response({ openFindings: [adminFinding], accountCount: 99 }));
    await flushPromises();

    // Both loads really happened; the stale one simply lost.
    expect(getEnvironmentAccounts).toHaveBeenCalledTimes(2);
    expect(wrapper.text()).not.toContain("New administrator account");
    expect(environment.value?.openAccountFindings).toBe(0);
  });

  // Submitting a left-over finding would send the previous environment's event
  // id with the new environment's id.
  it("drops a half-finished acknowledgement when the environment changes", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ openFindings: [adminFinding] }));
    const wrapper = await mountComponent();

    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Looks right")
      ?.trigger("click");
    await flushPromises();
    expect(bodyButton("Confirm")).toBeTruthy();

    routeParams.environmentId = "8";
    await flushPromises();

    expect(bodyButton("Confirm")).toBeFalsy();
  });

  // A successful response can still report an unsupported endpoint or a failed
  // read; without carrying the state the overview card keeps claiming ok.
  it("carries the monitoring state back to the shared environment", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ supported: false }));
    await mountComponent();
    expect(environment.value?.accountMonitoringState).toBe("unsupported");
  });

  it("reports an error state back to the shared environment", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ lastError: "missing frosh_tools:read" }));
    await mountComponent();
    expect(environment.value?.accountMonitoringState).toBe("error");
  });

  // The history is capped at the newest hundred events while open findings come
  // back in full, so a finding older than that cap would leave the filter empty
  // while the page plainly says something needs review.
  it("includes open findings older than the history cap in needs-review", async () => {
    getEnvironmentAccounts.mockResolvedValue(
      response({
        openFindings: [{ ...adminFinding, id: 500 }],
        events: [
          {
            id: 2,
            level: "info",
            kind: "user",
            change: "admin_revoked",
            label: "a***@e***.de",
            messageKey: "check.account.adminRevoked",
            params: { label: "a***@e***.de" },
            detectedAt: "2026-10-28T07:00:00Z",
          },
        ],
      }),
    );

    const wrapper = await mountComponent();
    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Needs review")
      ?.trigger("click");

    expect(wrapper.text()).toContain("New administrator account");
    expect(wrapper.text()).not.toContain("lost administrator rights");
  });

  // A scrape can refresh the finding with a worse state between the page
  // rendering and the click. Closing it on the older view would bury a change
  // nobody has read, so the backend rejects it and the reviewer is sent back.
  it("reloads and explains when the finding changed under the reviewer", async () => {
    getEnvironmentAccounts.mockResolvedValue(response({ openFindings: [adminFinding] }));
    acknowledgeEnvironmentAccountEvent.mockResolvedValue({
      data: undefined,
      error: { message: "account event changed since it was read" },
      response: new Response(null, { status: 409 }),
    });

    const wrapper = await mountComponent();
    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Looks right")
      ?.trigger("click");
    await flushPromises();
    bodyButton("Confirm")?.click();
    await flushPromises();

    expect(alerts.error).toHaveBeenCalledWith(
      "The finding changed while you were looking at it — please review the current state",
    );
    // Reloaded, so the reviewer sees the current state instead of the stale one.
    expect(getEnvironmentAccounts).toHaveBeenCalledTimes(2);
    expect(bodyButton("Confirm")).toBeFalsy();
  });
});
