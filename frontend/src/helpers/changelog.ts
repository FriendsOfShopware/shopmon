import { useLocale } from "@/composables/useLocale";

interface LocalizedChangelogEntry {
  text: string;
  textDe?: string | null;
}

/**
 * Returns a function that resolves an extension changelog entry's text for the
 * active UI locale, falling back to the English text when no German translation
 * is available.
 */
export function useChangelogText() {
  const { locale } = useLocale();

  return (entry: LocalizedChangelogEntry): string => {
    if (locale.value === "de" && entry.textDe) {
      return entry.textDe;
    }
    return entry.text;
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
