import * as Sentry from '@sentry/bun';
Sentry.init({
    dsn: process.env.SENTRY_DSN || '',
    release: process.env.SENTRY_RELEASE || 'unknown',
    environment: process.env.SENTRY_ENVIRONMENT || 'development',
    tracesSampleRate: 0.1,
    integrations: [Sentry.nativeNodeFetchIntegration()],
});
