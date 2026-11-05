import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";

import AccountMonitoringCard from "./AccountMonitoringCard.vue";

function mountCard(state: string, openFindings = 0) {
  return mount(AccountMonitoringCard, {
    props: { state, openFindings, to: { name: "account.environments.detail.accounts" } },
  });
}

describe("AccountMonitoringCard", () => {
  // Zero open findings is not automatically good news: these three states
  // produce zero too, and green would claim an all-clear nobody checked.
  it.each([
    ["pending", "Not initialised yet"],
    ["unsupported", "does not serve the account endpoint"],
    ["error", "last account check failed"],
  ])("warns instead of showing green for state %s", (state, expected) => {
    const wrapper = mountCard(state);
    expect(wrapper.text()).toContain(expected);
    expect(wrapper.text()).not.toContain("Nothing needs review");
  });

  it("shows the all-clear only once a baseline exists and the read succeeded", () => {
    expect(mountCard("ok").text()).toContain("Nothing needs review");
  });

  it("reports open findings over the all-clear", () => {
    const wrapper = mountCard("ok", 2);
    expect(wrapper.text()).toContain("2");
    expect(wrapper.text()).not.toContain("Nothing needs review");
  });

  it("says so when monitoring is switched off", () => {
    const wrapper = mountCard("off");
    expect(wrapper.text()).toContain("Account monitoring is switched off");
    expect(wrapper.text()).not.toContain("Nothing needs review");
  });

  // Open findings are stored, so a later failed or unsupported fetch does not
  // make them go away -- they still need review and must stay visible.
  it.each([
    ["error", "last account check failed"],
    ["unsupported", "does not serve the account endpoint"],
    ["pending", "Not initialised yet"],
  ])("keeps reporting open findings while %s", (state, warning) => {
    const wrapper = mountCard(state, 3);
    expect(wrapper.text()).toContain("3");
    expect(wrapper.text()).toContain(warning);
    expect(wrapper.text()).not.toContain("Nothing needs review");
  });
});
