import {
    adminClient,
    organizationClient,
} from 'better-auth/client/plugins';
import { passkeyClient } from '@better-auth/passkey/client'
import { ssoClient } from '@better-auth/sso/client'
import { createAuthClient } from 'better-auth/vue';

export const authClient = createAuthClient({
    basePath: '/auth',
    plugins: [
        adminClient(),
        passkeyClient(),
        organizationClient(),
        ssoClient(),
    ],
});
