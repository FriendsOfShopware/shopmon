import { afterEach, describe, expect, it, vi } from "vitest";

import { client } from "./generated/client.gen";
import { passkeyLogin, passkeyRegister } from "./generated";

describe("passkey client serialization", () => {
  afterEach(() => {
    client.setConfig({ fetch: undefined, baseUrl: "/api" });
  });

  it("sends passkey login as JSON, not a binary upload", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ token: "t", user: { id: "1", name: "n", email: "e" } }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    client.setConfig({ fetch: fetchMock, baseUrl: "http://example.test/api" });

    const body = { challengeKey: "abc", id: "cred", rawId: "cred", type: "public-key" };
    await passkeyLogin({ body });

    expect(fetchMock).toHaveBeenCalled();
    const request = fetchMock.mock.calls[0][0] as Request;
    expect(request.headers.get("Content-Type")).toContain("application/json");
    expect(JSON.parse(await request.clone().text())).toEqual(body);
  });

  it("sends passkey registration as JSON, not a binary upload", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ id: "pk-1", name: "Laptop" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    client.setConfig({ fetch: fetchMock, baseUrl: "http://example.test/api" });

    const body = {
      challengeKey: "abc",
      name: "Laptop",
      id: "cred",
      rawId: "cred",
      type: "public-key",
    };
    await passkeyRegister({ body });

    expect(fetchMock).toHaveBeenCalled();
    const request = fetchMock.mock.calls[0][0] as Request;
    expect(request.headers.get("Content-Type")).toContain("application/json");
    expect(JSON.parse(await request.clone().text())).toEqual(body);
  });
});
