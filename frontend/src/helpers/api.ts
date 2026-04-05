import createClient from "openapi-fetch";
import type { paths } from "../types/api";

export function getToken(): string | null {
  if (import.meta.env.SSR) return null;
  return localStorage.getItem("shopmon_token");
}

export function setToken(token: string | null) {
  if (import.meta.env.SSR) return;
  if (token) {
    localStorage.setItem("shopmon_token", token);
  } else {
    localStorage.removeItem("shopmon_token");
  }
}

export const api = createClient<paths>({
  baseUrl: "/api",
  headers: {},
});

// Add auth header to every request via middleware
api.use({
  async onRequest({ request }) {
    const token = getToken();
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`);
    }
    return request;
  },
});
