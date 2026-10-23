import { describe, expect, it } from "vitest";
import {
  advisoryOfficialId,
  isAdvisoryContentLocked,
  isLinkedShopmonAdvisory,
  isManualAdvisory,
  isShopmonAdvisory,
} from "./advisory";

describe("isShopmonAdvisory", () => {
  it("recognizes Shopmon-authored advisories", () => {
    expect(isShopmonAdvisory({ origin: "shopmon" })).toBe(true);
  });

  it("treats synced advisories as upstream", () => {
    expect(isShopmonAdvisory({ origin: "upstream" })).toBe(false);
    expect(isShopmonAdvisory({ origin: null })).toBe(false);
    expect(isShopmonAdvisory({})).toBe(false);
  });
});

describe("advisoryOfficialId", () => {
  it("prefers the CVE over the GHSA", () => {
    expect(advisoryOfficialId({ cve: "CVE-2026-1", ghsaId: "GHSA-a" })).toBe("CVE-2026-1");
  });

  it("falls back to the GHSA and then to null", () => {
    expect(advisoryOfficialId({ ghsaId: "GHSA-a" })).toBe("GHSA-a");
    expect(advisoryOfficialId({ cve: "", ghsaId: null })).toBeNull();
  });
});

describe("isManualAdvisory", () => {
  it("covers Shopmon-authored and GitHub-imported advisories", () => {
    expect(isManualAdvisory({ origin: "shopmon" })).toBe(true);
    expect(isManualAdvisory({ origin: "github" })).toBe(true);
    expect(isManualAdvisory({ origin: "upstream" })).toBe(false);
  });
});

describe("isAdvisoryContentLocked", () => {
  it("locks once a CVE is linked, not for a GHSA alone", () => {
    expect(isAdvisoryContentLocked({ origin: "github", ghsaId: "GHSA-a" })).toBe(false);
    expect(isAdvisoryContentLocked({ origin: "github", ghsaId: "GHSA-a", cve: "CVE-2026-1" })).toBe(
      true,
    );
    expect(isAdvisoryContentLocked({ origin: "shopmon" })).toBe(false);
  });
});

describe("isLinkedShopmonAdvisory", () => {
  it("is true only for Shopmon advisories with an official identifier", () => {
    expect(isLinkedShopmonAdvisory({ origin: "shopmon", cve: "CVE-2026-1" })).toBe(true);
    expect(isLinkedShopmonAdvisory({ origin: "shopmon" })).toBe(false);
    expect(isLinkedShopmonAdvisory({ origin: "upstream", cve: "CVE-2026-1" })).toBe(false);
  });
});
