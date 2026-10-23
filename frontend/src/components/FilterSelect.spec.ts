import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { enableAutoUnmount, flushPromises, mount } from "@vue/test-utils";
import FilterSelect from "./FilterSelect.vue";

enableAutoUnmount(afterEach);

describe("FilterSelect", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "ResizeObserver",
      class {
        observe() {}
        unobserve() {}
        disconnect() {}
      },
    );
    Element.prototype.scrollIntoView = vi.fn();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  async function setup() {
    const wrapper = mount(FilterSelect, {
      attachTo: document.body,
      props: {
        id: "package",
        modelValue: "all",
        label: "Package",
        searchable: true,
        searchPlaceholder: "Search packages…",
        emptyMessage: "No matching options.",
        options: [
          { value: "all", label: "All packages" },
          { value: "shopware/core", label: "shopware/core" },
          { value: "shopware/storefront", label: "shopware/storefront" },
          { value: "shopware/administration", label: "shopware/administration" },
        ],
        "onUpdate:modelValue": (value: string) => {
          void wrapper.setProps({ modelValue: value });
        },
      },
    });
    await wrapper.get("button").trigger("click");
    await flushPromises();
    const input = document.querySelector<HTMLInputElement>('input[role="combobox"]')!;
    return { wrapper, input };
  }

  async function search(input: HTMLInputElement, value: string) {
    input.value = value;
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await flushPromises();
  }

  it("uses a standard select without search by default", async () => {
    const wrapper = mount(FilterSelect, {
      attachTo: document.body,
      props: {
        id: "severity",
        modelValue: "all",
        label: "Severity",
        options: [
          { value: "all", label: "All severities" },
          { value: "high", label: "High" },
        ],
        "onUpdate:modelValue": (value: string) => {
          void wrapper.setProps({ modelValue: value });
        },
      },
    });
    await wrapper.get("button").trigger("keydown", { key: "ArrowDown" });
    await flushPromises();
    expect(document.querySelector("input")).toBeNull();
    const option = [...document.querySelectorAll<HTMLElement>('[role="option"]')].find(
      (item) => item.textContent?.trim() === "High",
    )!;
    expect(option).toBeDefined();
    option.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
    );
    await flushPromises();
    expect(wrapper.emitted("update:modelValue")).toEqual([["high"]]);
    expect(wrapper.get("button").text()).toContain("High");
    expect(wrapper.get("button").attributes("aria-expanded")).toBe("false");
  });

  it("filters inline without changing the selection, then selects and closes", async () => {
    const { wrapper, input } = await setup();
    await search(input, "STORE");
    const options = document.querySelectorAll<HTMLElement>('[role="option"]');
    expect([...options].map((option) => option.textContent?.trim())).toEqual([
      "shopware/storefront",
    ]);
    expect(wrapper.emitted("update:modelValue")).toBeUndefined();
    options[0].click();
    await flushPromises();
    expect(wrapper.emitted("update:modelValue")).toEqual([["shopware/storefront"]]);
    expect(wrapper.get("button").text()).toContain("shopware/storefront");
    expect(wrapper.get("button").attributes("aria-expanded")).toBe("false");
  });

  it("shows no matches and restores all options when reopened", async () => {
    const { wrapper, input } = await setup();
    await search(input, "not-a-package");
    expect(document.body.textContent).toContain("No matching options.");
    expect(document.querySelectorAll('[role="option"]')).toHaveLength(0);
    input.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    await flushPromises();
    expect(wrapper.emitted("update:modelValue")).toBeUndefined();
    await wrapper.get("button").trigger("click");
    await flushPromises();
    expect(document.querySelectorAll('[role="option"]')).toHaveLength(4);
    expect(document.querySelector<HTMLInputElement>("input")!.value).toBe("");
  });

  it("selects a filtered option with the keyboard", async () => {
    const { wrapper, input } = await setup();
    await search(input, "core");
    input.dispatchEvent(
      new KeyboardEvent("keydown", { key: "ArrowDown", bubbles: true, cancelable: true }),
    );
    await flushPromises();
    input.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
    );
    await flushPromises();
    expect(wrapper.emitted("update:modelValue")).toEqual([["shopware/core"]]);
  });
});
