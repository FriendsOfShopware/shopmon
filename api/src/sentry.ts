import * as Sentry from '@sentry/node';

if (process.env.SENTRY_DSN) {
    Sentry.init({
        dsn: process.env.SENTRY_DSN || '',
        release: process.env.SENTRY_RELEASE || 'unknown',
        environment: process.env.SENTRY_ENVIRONMENT || 'development',
        tracesSampleRate: 0.1,
        integrations: [Sentry.nativeNodeFetchIntegration()],
        _experiments: {
            enableLogs: true,
        },
    });
}
