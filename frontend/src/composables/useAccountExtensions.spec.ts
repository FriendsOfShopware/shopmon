import { describe, expect, it } from "vitest";
import {
  type AccountExtension,
  type AccountExtensionEnvironment,
  envHasUpdate,
  hasUpdate,
  updateCount,
} from "./useAccountExtensions";

function env(overrides: Partial<AccountExtensionEnvironment> = {}): AccountExtensionEnvironment {
  return {
    environmentId: 1,
    environmentName: "prod",
    environmentShopName: "Shop",
    environmentOrganizationName: "org",
    environmentOrganizationId: "org-1",
    environmentUrl: "https://example.com",
    shopId: 1,
    shopName: "Shop",
    active: true,
    version: "1.0.0",
    latestVersion: "1.0.0",
    installed: true,
    ...overrides,
  };
}

function ext(overrides: Partial<AccountExtension> = {}): AccountExtension {
  return {
    name: "SwagExample",
    label: "Example",
    version: "1.0.0",
    latestVersion: "1.0.0",
    active: true,
    installed: true,
    environments: [],
    ...overrides,
  } as AccountExtension;
}

describe("envHasUpdate", () => {
  it("is true when an installed environment runs behind the latest version", () => {
    expect(envHasUpdate(env({ version: "1.0.0", latestVersion: "1.1.0" }))).toBe(true);
  });

  it("is false when up to date or not installed", () => {
    expect(envHasUpdate(env({ version: "1.1.0", latestVersion: "1.1.0" }))).toBe(false);
    expect(envHasUpdate(env({ version: "1.0.0", latestVersion: "1.1.0", installed: false }))).toBe(
      false,
    );
  });
});

describe("hasUpdate", () => {
  it("is true when any environment runs an outdated version", () => {
    const extension = ext({
      // Aggregate top-level fields report the extension as up to date...
      version: "1.1.0",
      latestVersion: "1.1.0",
      environments: [
        env({ environmentId: 1, version: "1.1.0", latestVersion: "1.1.0" }),
        // ...but a second environment is still behind and must be counted.
        env({ environmentId: 2, version: "1.0.0", latestVersion: "1.1.0" }),
      ],
    });

    expect(hasUpdate(extension)).toBe(true);
  });

  it("is false when every environment is up to date", () => {
    const extension = ext({
      environments: [
        env({ environmentId: 1, version: "1.1.0", latestVersion: "1.1.0" }),
        env({ environmentId: 2, version: "1.1.0", latestVersion: "1.1.0" }),
      ],
    });

    expect(hasUpdate(extension)).toBe(false);
  });

  it("is false when the environments list is empty", () => {
    // ext() defaults to environments: [] — no environment to check means no update.
    expect(hasUpdate(ext())).toBe(false);
  });
});

describe("updateCount", () => {
  it("counts every environment that has an update available", () => {
    const extension = ext({
      environments: [
        env({ environmentId: 1, version: "1.0.0", latestVersion: "1.1.0" }),
        env({ environmentId: 2, version: "1.1.0", latestVersion: "1.1.0" }),
        env({ environmentId: 3, version: "1.0.0", latestVersion: "1.1.0" }),
      ],
    });

    expect(updateCount(extension)).toBe(2);
  });
});
