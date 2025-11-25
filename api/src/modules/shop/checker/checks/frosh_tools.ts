import type { Checker, CheckerInput, CheckerOutput } from '../registery.ts';

const ignores = [
    'frosh-tools.checker.scheduledTaskGood',
    'frosh-tools.checker.scheduledTaskWarning',
    'frosh-tools.checker.queuesGood',
    'frosh-tools.checker.queuesWarning',
    'frosh-tools.checker.prodGood',
    'frosh-tools.checker.not-prod',
    'frosh-tools.checker.adminWorkerGood',
    'frosh-tools.checker.adminWorkerWarning',
];

const snippets: Record<string, string> = {
    'frosh-tools.checker.mysqlError': 'MySQL Version cannot be checked',
    'frosh-tools.checker.mysqlDbVersion': 'MySQL Version',
    'frosh-tools.checker.mysqlDbVersionError':
        'MySQL Version has technical problems',
    'frosh-tools.checker.mysqlDbOutdated': 'MySQL Version is outdated',
    'frosh-tools.checker.mariaDbVersion': 'MariaDB Version',
    'frosh-tools.checker.phpOutdated': 'PHP Version is outdated',
    'frosh-tools.checker.phpGood': 'PHP Version',
    'frosh-tools.checker.maxExecutionTimeError':
        'Max-Execution-Time is too low',
    'frosh-tools.checker.maxExecutionTimeGood': 'Max-Execution-Time',
    'frosh-tools.checker.memoryLimitError': 'Memory-Limit is too low',
    'frosh-tools.checker.memoryLimitGood': 'Memory-Limit',
    'frosh-tools.checker.zendOpcacheGood': 'Zend Opcache is active',
    'frosh-tools.checker.zendOpcacheWarning': 'Zend Opcache is not active',
    'frosh-tools.checker.esGood': 'Elasticsearch is enabled',
    'frosh-tools.checker.esWarning': 'Elasticsearch is disabled',
    'frosh-tools.checker.adminWorkerGood': 'Admin-Worker is disabled',
    'frosh-tools.checker.adminWorkerWarning': 'Admin-Worker should be disabled',
    'frosh-tools.checker.publicFilesystemGood': 'PublicFilesystem is not local',
    'frosh-tools.checker.publicFilesystemWarning':
        'PublicFilesystem should not be local',
    'frosh-tools.checker.updateMailVarialesGood':
        'MailVariables are not updated frequently',
    'frosh-tools.checker.updateMailVarialesWarning':
        'MailVariables should not be updated frequently',
    'frosh-tools.checker.mailNotSendWithQueue':
        'Mails should be sent using the message queue',
    'frosh-tools.checker.mailSendWithQueue':
        'Mails are send with the message queue',
    'frosh-tools.checker.incrementStorageIsDB':
        'Increment storage is heavily using the Storage. This feature should be disabled or Redis should be used',
    'frosh-tools.checker.incrementStorageIsNotDB':
        'Increment storage is correct configured',
    'frosh-tools.checker.queueIsDefault':
        'The default queue storage in database does not scale well with multiple workers',
    'frosh-tools.checker.queueIsOk':
        'Configured queue storage is ok for multiple workers',
    'frosh-tools.checker.fixCacheIdNotSet': 'A fixed cache id should be set',
    'frosh-tools.checker.fixCacheIdIsSet': 'A fixed cache id is set',
    'frosh-tools.checker.BusinessEventHandlerLevelGood':
        'BusinessEventHandler does not log infos',
    'frosh-tools.checker.BusinessEventHandlerLevelWarning':
        'BusinessEventHandler is logging infos',
    'frosh-tools.checker.AssertActiveGood':
        'PHP value assert.active is disabled',
    'frosh-tools.checker.AssertActiveWarning':
        'PHP value assert.active is not disabled',
    'frosh-tools.checker.EnableFileOverrideGood':
        'PHP value opcache.enable_file_override is enabled',
    'frosh-tools.checker.EnableFileOverrideWarning':
        'PHP value opcache.enable_file_override is not enabled',
    'frosh-tools.checker.InternedStringsBufferGood':
        'PHP value opcache.interned_strings_buffer has minimum value',
    'frosh-tools.checker.InternedStringsBufferWarning':
        'PHP value opcache.interned_strings_buffer is too low',
    'frosh-tools.checker.ZendDetectUnicodeGood':
        'PHP value zend.detect_unicode is disabled',
    'frosh-tools.checker.ZendDetectUnicodeWarning':
        'PHP value zend.detect_unicode is not disabled',
    'frosh-tools.checker.RealpathCacheTtlGood':
        'PHP value realpath_cache_ttl is good',
    'frosh-tools.checker.RealpathCacheTtlWarning':
        'PHP value realpath_cache_ttl is low',
};

interface FroshToolsCheck {
    id?: string;
    state: 'STATE_OK' | 'STATE_WARNING' | 'STATE_ERROR';
    snippet: string;
    current: string;
    recommended: string;
    url: string;
}

export default class implements Checker {
    async check(input: CheckerInput, result: CheckerOutput): Promise<void> {
        for (const extension of input.extensions) {
            if (
                extension.installed &&
                extension.active &&
                extension.name === 'FroshTools'
            ) {
                try {
                    const healthStatus = await input.client.get(
                        '/_action/frosh-tools/health/status',
                    );
                    const performanceStatus = await input.client.get(
                        '/_action/frosh-tools/performance/status',
                    );

                    mapToShopmon(
                        healthStatus.body as FroshToolsCheck[],
                        result,
                    );
                    mapToShopmon(
                        performanceStatus.body as FroshToolsCheck[],
                        result,
                    );
                } catch (e) {
                    result.error(
                        'frosh.tools.connect',
                        `FroshTools is not reachable: Error: ${e}`,
                    );
                    return;
                }
            }
        }
    }
}

function mapToShopmon(checks: FroshToolsCheck[], result: CheckerOutput) {
    for (const status of checks) {
        if (ignores.includes(status.id || status.snippet)) {
            continue;
        }

        let msg = snippets[status.snippet] || status.snippet;

        if (status.current) {
            msg += ` Current: ${status.current}, Recommended: ${status.recommended}`;
        }

        if (status.state === 'STATE_OK') {
            result.success(
                status.id || status.snippet,
                msg,
                'FroshTools',
                status.url,
            );
        } else if (status.state === 'STATE_WARNING') {
            result.warning(
                status.id || status.snippet,
                msg,
                'FroshTools',
                status.url,
            );
        } else if (status.state === 'STATE_ERROR') {
            result.error(
                status.id || status.snippet,
                msg,
                'FroshTools',
                status.url,
            );
        }
    }
}
