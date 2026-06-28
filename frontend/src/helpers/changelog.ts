import { sanitizeHtml } from "@/helpers/sanitize";

interface ChangelogEntry {
  text: string;
}

// Returns an extension changelog entry's text, sanitized for v-html rendering.
// The text is already resolved to the requested language by the API (the
// `language` query parameter, with English fallback); this only strips the
// untrusted store HTML of any XSS vectors.
export function useChangelogText() {
  return (entry: ChangelogEntry): string => sanitizeHtml(entry.text);
}

export function sumChanges(changes: {
  oldShopwareVersion?: string | null;
  newShopwareVersion?: string | null;
  extensions: { state: string }[];
}) {
  const messages: string[] = [];

  if (changes.oldShopwareVersion && changes.newShopwareVersion) {
    messages.push(
      `Shopware Update from ${changes.oldShopwareVersion} to ${changes.newShopwareVersion}`,
    );
  }

  const stateCounts: Record<string, number> = {};
  for (const extension of changes.extensions) {
    if (stateCounts[extension.state] !== undefined) {
      stateCounts[extension.state] = stateCounts[extension.state] + 1;
    } else {
      stateCounts[extension.state] = 1;
    }
  }

  for (const [state, count] of Object.entries(stateCounts)) {
    messages.push(`${state} ${count} extension${count > 1 ? "s" : ""}`);
  }

  return messages.join(", ");
}
