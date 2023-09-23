import type { ShopChangelog } from '@apiTypes/shop'; 
import type { Changelogs } from '@apiTypes/dashboard'; 

export function sumChanges(changes: ShopChangelog | Changelogs) {
    const messages: string[] = [];
  
    if (changes.old_shopware_version && changes.new_shopware_version) {
      messages.push(
        `Shopware Update from ${changes.old_shopware_version} to ${changes.new_shopware_version}`
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
