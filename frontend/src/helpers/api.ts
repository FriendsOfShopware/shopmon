import createClient from "openapi-fetch";
import type { paths } from "../types/api";

const TOKEN_KEY = "shopmon_token";
// Saved while an admin is impersonating so stop-impersonating can restore the
// original session after the short-lived impersonation token is deleted.
const ADMIN_TOKEN_KEY = "shopmon_admin_token";

export function getToken(): string | null {
  if (import.meta.env.SSR) return null;
  return localStorage.getItem(TOKEN_KEY);
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
    localStorage.setItem(TOKEN_KEY, token);
  } else {
    localStorage.removeItem(TOKEN_KEY);
  }
}

/** Stash the current session token before switching to an impersonation token. */
export function stashAdminToken(): void {
  if (import.meta.env.SSR) return;
  const token = getToken();
  if (token) {
    localStorage.setItem(ADMIN_TOKEN_KEY, token);
  }
}

/**
 * Restore the admin session after stop-impersonating.
 * Returns the restored token, or null if none was stashed.
 */
export function restoreAdminToken(): string | null {
  if (import.meta.env.SSR) return null;
  const token = localStorage.getItem(ADMIN_TOKEN_KEY);
  localStorage.removeItem(ADMIN_TOKEN_KEY);
  if (token) {
    setToken(token);
  }
  return token;
}

export function clearAdminToken(): void {
  if (import.meta.env.SSR) return;
  localStorage.removeItem(ADMIN_TOKEN_KEY);
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
