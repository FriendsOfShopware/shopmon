import { describe, expect, it } from "vitest";
import { shopwareStoreSearchUrl } from "./shopware-store";

describe("shopwareStoreSearchUrl", () => {
  it("builds a Shopware Store search URL for an extension name", () => {
    expect(shopwareStoreSearchUrl("Swag PayPal")).toBe(
      "https://store.shopware.com/en/search?search=Swag%20PayPal",
    );
  });
});
