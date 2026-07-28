import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getAccountShops: vi.fn(),
}));

import { getAccountShops } from "@/api/generated";
import { fetchAccountShops, resetAccountShops, useAccountShops } from "./useAccountShops";

describe("useAccountShops", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetAccountShops();
  });

  it("fetches shops once and shares state", async () => {
    const shops = [{ id: 1, name: "Shop A" }];
    vi.mocked(getAccountShops).mockResolvedValue({
      data: shops,
      error: undefined,
      response: new Response(),
    } as never);

    const first = useAccountShops();
    const second = useAccountShops();

    await fetchAccountShops();

    expect(getAccountShops).toHaveBeenCalledTimes(1);
    expect(first.shops.value).toEqual(shops);
    expect(second.shops.value).toEqual(shops);
  });

  it("re-fetches after reset", async () => {
    vi.mocked(getAccountShops)
      .mockResolvedValueOnce({
        data: [{ id: 1, name: "Shop A" }],
        error: undefined,
        response: new Response(),
      } as never)
      .mockResolvedValueOnce({
        data: [{ id: 2, name: "Shop B" }],
        error: undefined,
        response: new Response(),
      } as never);

    await fetchAccountShops();
    resetAccountShops();
    await fetchAccountShops();

    expect(getAccountShops).toHaveBeenCalledTimes(2);
    expect(useAccountShops().shops.value).toEqual([{ id: 2, name: "Shop B" }]);
  });
});
