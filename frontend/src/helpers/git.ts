const COMMIT_PLACEHOLDER_FALLBACK = "0123456789abcdef0123456789abcdef01234567";

export const GIT_COMMIT_PLACEHOLDER = "%commit%";
export const GIT_URL_PLACEHOLDER = "https://github.com/your-org/your-repo/commit/%commit%";

function normalizeGitUrl(url: string) {
  return url.replaceAll(GIT_COMMIT_PLACEHOLDER, COMMIT_PLACEHOLDER_FALLBACK);
}

export function isValidGitUrl(url: string) {
  try {
    new URL(normalizeGitUrl(url));
    return true;
  } catch {
    return false;
  }
}
