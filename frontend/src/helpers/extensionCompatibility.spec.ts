import { describe, expect, it } from "vitest";
import { compareByCompatibilitySeverity, getCompatibilityState } from "./extensionCompatibility";

describe("getCompatibilityState", () => {
  it("returns inactive for inactive extensions regardless of compatibility", () => {
    expect(getCompatibilityState({ active: false })).toBe("inactive");
    expect(
      getCompatibilityState({
        active: false,
        compatibility: { label: "y", type: "red" },
      }),
    ).toBe("inactive");
  });

  it("returns unknown for active extensions without compatibility info", () => {
    expect(getCompatibilityState({ active: true })).toBe("unknown");
    expect(getCompatibilityState({ active: true, compatibility: null })).toBe("unknown");
  });

  it("returns incompatible for red compatibility type", () => {
    expect(
      getCompatibilityState({
        active: true,
        compatibility: { label: "Not compatible", type: "red" },
      }),
    ).toBe("incompatible");
  });

  it("returns updateAvailable when a compatible update is available now", () => {
    expect(
      getCompatibilityState({
        active: true,
        compatibility: { label: "Available now", type: "green" },
      }),
    ).toBe("updateAvailable");
  });

  it("returns compatible otherwise", () => {
    expect(
      getCompatibilityState({
        active: true,
        compatibility: { label: "Compatible", type: "green" },
      }),
    ).toBe("compatible");
  });
});

describe("compareByCompatibilitySeverity", () => {
  it("orders incompatible before unknown, updateAvailable, compatible and inactive", () => {
    const extensions = [
      { active: false },
      { active: true, compatibility: { label: "Compatible", type: "green" } },
      { active: true },
      { active: true, compatibility: { label: "Available now", type: "green" } },
      { active: true, compatibility: { label: "Broken", type: "red" } },
    ];

    const sorted = [...extensions].sort(compareByCompatibilitySeverity);

    expect(sorted.map(getCompatibilityState)).toEqual([
      "incompatible",
      "unknown",
      "updateAvailable",
      "compatible",
      "inactive",
    ]);
  });
});
