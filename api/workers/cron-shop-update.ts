import {onSchedule} from '../src/cron/schedule';
import { createSentry } from '../src/sentry';

export default {
    scheduled(event: ScheduledEvent, env: Env, ctx: ExecutionContext) {
        const sentry = createSentry(ctx, env);

        ctx.waitUntil(
            onSchedule(env)
            .catch(err => sentry.captureException(err))
        )
    }
}