import { describe, expect, it } from "vitest";
import { isNewerExtensionVersion } from "./extensionVersion";

describe("isNewerExtensionVersion", () => {
  it("only identifies a newer comparable release as an update", () => {
    expect(isNewerExtensionVersion("2.8.1", "2.8.0")).toBe(false);
    expect(isNewerExtensionVersion("2.8.0", "2.8.1")).toBe(true);
    expect(isNewerExtensionVersion("2.8.1", "2.8.1")).toBe(false);
    expect(isNewerExtensionVersion("2.9.0", "2.10.0")).toBe(true);
    expect(isNewerExtensionVersion("1.0.0-rc.1", "1.0.0")).toBe(true);
    expect(isNewerExtensionVersion("dev-main", "2.8.1")).toBe(false);
    expect(isNewerExtensionVersion("2.8.1", null)).toBe(false);
  });
});
