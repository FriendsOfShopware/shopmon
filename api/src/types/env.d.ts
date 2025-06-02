declare global {
    namespace NodeJS {
        interface ProcessEnv {
            APP_DATABASE_PATH: string;
            APP_SECRET: string;

            // Email
            MAIL_ACTIVE?: 'true' | 'false';
            MAIL_FROM: string;

            // SMTP Configuration
            SMTP_HOST?: string;
            SMTP_PORT?: string;
            SMTP_SECURE?: 'true' | 'false';
            SMTP_USER?: string;
            SMTP_PASS?: string;

            APP_OAUTH_GITHUB_CLIENT_ID: string;
            APP_OAUTH_GITHUB_CLIENT_SECRET: string;

            // External APIs
            PAGESPEED_API_KEY?: string;

            // Application
            FRONTEND_URL?: string;
            DISABLE_REGISTRATION?: string;
            APP_FILES_DIR?: string;

            // Monitoring (optional)
            SENTRY_DSN?: string;
            SENTRY_RELEASE?: string;
        }
    }
}

// Make this file a module
export {};
