import type { Checker, CheckerInput, CheckerOutput } from '../registery.ts';

export default class implements Checker {
    async check(input: CheckerInput, result: CheckerOutput): Promise<void> {
        const securityInfo = await getSecurityInfo();

        if (
            securityInfo.versionToAdvisories[input.config.version] === undefined
        ) {
            return;
        }

        for (const extension of input.extensions) {
            if (
                extension.name === 'SwagPlatformSecurity' &&
                extension.active &&
                extension.version === extension.latestVersion
            ) {
                return; // User has security plugin
            }
        }

        for (const advisoryId of securityInfo.versionToAdvisories[
            input.config.version
        ]) {
            const advisory = securityInfo.advisories[advisoryId];

            result.error(
                `advisory.${advisory.cve}`,
                `Security Issue: ${advisory.title}`,
                advisory.source,
                advisory.link,
            );
        }
    }
}

async function getSecurityInfo(): Promise<Security> {
    const securityInfo = await fetch(
        'https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/security.json',
    );

    return (await securityInfo.json()) as Security;
}

interface Security {
    latestPluginVersion: string;
    advisories: Record<string, Advisory>;
    versionToAdvisories: Record<string, string[]>;
}

interface Advisory {
    title: string;
    link: string;
    cve: string;
    affectedVersions: string;
    source: string;
    reportedAt: string;
}
