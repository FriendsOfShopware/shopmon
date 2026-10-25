/** Turn resolved client errors into failures before applying local changes. */
export function requireMutation(result: { error?: unknown }, fallback: string): void {
  if (result.error !== undefined) {
    throw new Error(apiErrorMessage(result.error, fallback));
  }
}

export function apiErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === "string" && error.trim()) return error;
  if (
    error &&
    typeof error === "object" &&
    "message" in error &&
    typeof error.message === "string" &&
    error.message.trim()
  )
    return error.message;
  return fallback;
}
