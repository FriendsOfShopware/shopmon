import createClient from "openapi-fetch";
import type { paths } from "../types/api";

export function getToken(): string | null {
  if (import.meta.env.SSR) return null;
  return localStorage.getItem("shopmon_token");
}

// The supported API content languages. Server-side localized store text resolves
// to one of these (falling back to English), so the active UI locale is mapped
// down to this set before being sent as the `language` query parameter.
export type ApiLanguage = "en" | "de";

// Returns the language to request localized API content in, derived from the
// persisted UI locale (same key useLocale writes). Defaults to English.
export function apiLanguage(): ApiLanguage {
  if (import.meta.env.SSR) return "en";
  return localStorage.getItem("shopmon-locale") === "de" ? "de" : "en";
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
