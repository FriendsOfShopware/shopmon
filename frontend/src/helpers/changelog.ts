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
        messages.push(`${state} ${count} extension${count > 1 ? 's' : ''}`);
    }

    return messages.join(', ');
}
