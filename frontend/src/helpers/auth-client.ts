// This module re-exports the new API client and session composable
// to ease migration from better-auth. All auth operations now go
// through the openapi-fetch based `api` client.
export { api } from "./api";
export { useSession, fetchSession } from "../composables/useSession";
export type { SessionData } from "../composables/useSession";
