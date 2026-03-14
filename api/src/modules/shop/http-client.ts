const HEADER_NAME = "shopmon-shop-token";

export class HttpClientResponse<ResponseType> {
  statusCode: number;
  body: ResponseType;
  headers: Headers;
  constructor(statusCode: number, body: ResponseType, headers: Headers) {
    this.statusCode = statusCode;
    this.body = body;
    this.headers = headers;
  }
}

export class ApiClientAuthenticationFailed extends Error {
  response: HttpClientResponse<string>;
  constructor(response: HttpClientResponse<string>) {
    super(
      `Authentication failed with status ${response.statusCode}: ${JSON.stringify(response.body)}`,
    );
    this.response = response;
  }
}

export class ApiClientRequestFailed extends Error {
  response: HttpClientResponse<ShopwareErrorResponse>;
  constructor(response: HttpClientResponse<ShopwareErrorResponse>) {
    const message = response.body.errors?.map((e) => e.detail).join(", ") ?? "Unknown error";
    super(`Request failed: ${message}`);
    this.response = response;
  }
}

type ShopwareErrorResponse = {
  errors: {
    code: string;
    status: string;
    title: string;
    detail: string;
  }[];
};

interface TokenCacheItem {
  token: string;
  expiresIn: Date;
}

interface ShopCredentials {
  url: string;
  clientId: string;
  clientSecret: string;
  shopToken: string;
}

export class HttpClient {
  private shop: ShopCredentials;
  private cachedToken: TokenCacheItem | null = null;

  constructor(shop: ShopCredentials) {
    this.shop = shop;
  }

  async get<R>(url: string, headers: Record<string, string> = {}): Promise<HttpClientResponse<R>> {
    return this.request("GET", url, null, headers);
  }

  async post<R>(
    url: string,
    json: object = {},
    headers: Record<string, string> = {},
  ): Promise<HttpClientResponse<R>> {
    headers["content-type"] = "application/json";
    headers.accept = "application/json";
    return this.request("POST", url, JSON.stringify(json), headers);
  }

  async put<R>(
    url: string,
    json: object = {},
    headers: Record<string, string> = {},
  ): Promise<HttpClientResponse<R>> {
    headers["content-type"] = "application/json";
    headers.accept = "application/json";
    return this.request("PUT", url, JSON.stringify(json), headers);
  }

  async patch<R>(
    url: string,
    json: object = {},
    headers: Record<string, string> = {},
  ): Promise<HttpClientResponse<R>> {
    headers["content-type"] = "application/json";
    headers.accept = "application/json";
    return this.request("PATCH", url, JSON.stringify(json), headers);
  }

  async delete<R>(
    url: string,
    json: object = {},
    headers: Record<string, string> = {},
  ): Promise<HttpClientResponse<R>> {
    headers["content-type"] = "application/json";
    headers.accept = "application/json";
    return this.request("DELETE", url, JSON.stringify(json), headers);
  }

  async getToken(): Promise<string> {
    if (this.cachedToken && this.cachedToken.expiresIn.getTime() > Date.now()) {
      return this.cachedToken.token;
    }

    this.cachedToken = null;

    const resp = await globalThis.fetch(this.buildUrl(this.shop.url, "/api/oauth/token"), {
      method: "POST",
      redirect: "manual",
      headers: {
        "content-type": "application/json",
        [HEADER_NAME]: this.shop.shopToken,
      },
      body: JSON.stringify({
        grant_type: "client_credentials",
        client_id: this.shop.clientId,
        client_secret: this.shop.clientSecret,
      }),
    });

    if (resp.status === 301 || resp.status === 302) {
      throw new ApiClientRequestFailed(
        new HttpClientResponse(
          resp.status,
          {
            errors: [
              {
                code: "301",
                status: "301",
                title: "Redirect",
                detail:
                  "Got a redirect response, the URL should point to the Shop without redirect",
              },
            ],
          },
          resp.headers,
        ),
      );
    }

    if (!resp.ok) {
      const contentType = resp.headers.get("content-type") ?? "text/plain";
      const body = contentType.includes("application/json") ? await resp.json() : await resp.text();
      throw new ApiClientAuthenticationFailed(
        new HttpClientResponse(resp.status, body as string, resp.headers),
      );
    }

    const authBody = (await resp.json()) as { access_token: string; expires_in: number };
    const expireDate = new Date();
    expireDate.setSeconds(expireDate.getSeconds() + authBody.expires_in);

    this.cachedToken = { token: authBody.access_token, expiresIn: expireDate };
    return this.cachedToken.token;
  }

  private buildUrl(...parts: string[]): string {
    return parts.map((part) => part.replace(/^\/+|\/+$/g, "")).join("/");
  }

  private async request<R>(
    method: string,
    url: string,
    body: string | null,
    headers: Record<string, string>,
  ): Promise<HttpClientResponse<R>> {
    const f = await globalThis.fetch(this.buildUrl(this.shop.url, "/api", url), {
      redirect: "manual",
      body,
      headers: {
        Authorization: `Bearer ${await this.getToken()}`,
        [HEADER_NAME]: this.shop.shopToken,
        ...headers,
      },
      method,
    });

    if (f.status === 301 || f.status === 302) {
      throw new ApiClientRequestFailed(
        new HttpClientResponse(
          f.status,
          {
            errors: [
              {
                code: "301",
                status: "301",
                title: "Redirect",
                detail:
                  "Got a redirect response, the URL should point to the Shop without redirect",
              },
            ],
          },
          f.headers,
        ),
      );
    }

    if (!f.ok && f.status === 401) {
      this.cachedToken = null;
      return this.request(method, url, body, headers);
    }

    if (!f.ok) {
      throw new ApiClientRequestFailed(
        new HttpClientResponse(f.status, (await f.json()) as ShopwareErrorResponse, f.headers),
      );
    }

    if (f.status === 204) {
      return new HttpClientResponse(f.status, {} as R, f.headers);
    }

    return new HttpClientResponse(f.status, (await f.json()) as R, f.headers);
  }
}
