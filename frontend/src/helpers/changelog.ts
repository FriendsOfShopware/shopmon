import type { ShopChangelog } from '@/types/shop';
import type { Changelogs } from '@/types/dashboard';

export function sumChanges(changes: ShopChangelog | Changelogs) {
  const messages: string[] = [];

  if (changes.oldShopwareVersion && changes.newShopwareVersion) {
    messages.push(
      `Shopware Update from ${changes.oldShopwareVersion} to ${changes.newShopwareVersion}`
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
    messages.push(
      `${state} ${count} extension` + (count > 1 ? 's' : '')
    )
  }

  return messages.join(', ');
}
