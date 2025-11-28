import type { Checker, CheckerInput, CheckerOutput } from '../registery.ts';

export default class implements Checker {
    VAILD_ENVIRONMENTS = ['production', 'staging', 'prod', 'stage'];

    async check(input: CheckerInput, result: CheckerOutput): Promise<void> {
        if (!this.VAILD_ENVIRONMENTS.includes(input.cacheInfo.environment)) {
            result.warning(
                'shopware.env',
                `Environment is not set to production or staging. It is set to ${input.cacheInfo.environment}`,
            );
            return;
        }

        result.success(
            'shopware.env',
            `Environment is set correctly to ${input.cacheInfo.environment}`,
        );
    }
}
