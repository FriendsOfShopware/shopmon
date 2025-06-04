<template>
    <div class="login-header">
        <h2>
            Confirming your Account Registration
        </h2>
    </div>
    
    <Alert v-if="isLoading" type="info">
        Loading...
    </Alert>

    <template v-else>
        <Alert v-if="success" type="success">
            Your email address has been confirmed.
            <router-link :to="{ name: 'account.login' }">
                Login
            </router-link>
        </Alert>

        <Alert v-else type="error">
            The given token has been expired
        </Alert>
    </template>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { authClient } from '@/helpers/auth-client';

const alert = useAlert();
</script>

<script lang="ts">
export default {
    data() {
        return {
            isLoading: true,
            success: false,
        };
    },
    async created() {
        const resp = await authClient.verifyEmail({query: {token: this.$route.params.token as string}});

        if (resp.error) {
            this.success = false;
            this.isLoading = false;

            alert.error(resp.error.message || 'Failed to verify email');

            return;
        }

        this.success = true;
        this.isLoading = false;
    },
    methods: {
        goToLogin() {
            this.$router.push({ name: 'account.login' });
        },
    },
};
</script>
