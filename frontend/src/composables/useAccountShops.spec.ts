import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/helpers/api", () => ({
  api: {
    GET: vi.fn(),
  },
}));

import { api } from "@/helpers/api";
import { fetchAccountShops, resetAccountShops, useAccountShops } from "./useAccountShops";

describe("useAccountShops", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetAccountShops();
  });

  it("fetches shops once and shares state", async () => {
    const shops = [{ id: 1, name: "Shop A" }];
    vi.mocked(api.GET).mockResolvedValue({
      data: shops,
      error: undefined,
      response: new Response(),
    } as never);

    const first = useAccountShops();
    const second = useAccountShops();

    await fetchAccountShops();

    expect(api.GET).toHaveBeenCalledTimes(1);
    expect(api.GET).toHaveBeenCalledWith("/account/shops");
    expect(first.shops.value).toEqual(shops);
    expect(second.shops.value).toEqual(shops);
  });

  it("re-fetches after reset", async () => {
    vi.mocked(api.GET)
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

    expect(api.GET).toHaveBeenCalledTimes(2);
    expect(useAccountShops().shops.value).toEqual([{ id: 2, name: "Shop B" }]);
  });
});
