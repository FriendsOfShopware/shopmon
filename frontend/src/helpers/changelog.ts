import { useLocale } from "@/composables/useLocale";
import { sanitizeHtml } from "@/helpers/sanitize";

interface ChangelogEntry {
  text: string;
  textDe?: string | null;
}

// Returns an extension changelog entry's text, sanitized for v-html rendering.
// Store catalog endpoints already resolve `text` to the requested language, so
// textDe is absent there and the fallback simply returns `text`. Stored
// environment-changelog history still carries both languages (text = en_GB,
// textDe = de_DE), so honour the active locale for those entries.
export function useChangelogText() {
  const { locale } = useLocale();

  return (entry: ChangelogEntry): string => {
    if (locale.value === "de" && entry.textDe) {
      return sanitizeHtml(entry.textDe);
    }
    return sanitizeHtml(entry.text);
  };
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
