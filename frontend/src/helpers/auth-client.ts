import {
    adminClient,
    organizationClient,
    passkeyClient,
} from 'better-auth/client/plugins';
import { createAuthClient } from 'better-auth/vue';

export const authClient = createAuthClient({
    basePath: '/auth',
    plugins: [adminClient(), passkeyClient(), organizationClient()],
});
