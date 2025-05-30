<template>
    <div class="login-header">
        <h2>
            Confirming your Account Registration
        </h2>
    </div>
    
    <Alert type="info" v-if="isLoading">
        Loading...
    </Alert>

    <template v-else>
        <Alert type="success" v-if="success">
            Your email address has been confirmed.
            <router-link :to="{ name: 'account.login' }">
                Login
            </router-link>
        </Alert>

        <Alert type="error" v-else>
            The given token has been expired
        </Alert>
    </template>
</template>

<script setup lang="ts">
import { authClient } from '@/helpers/auth-client';
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
        try {
            await authClient.verifyEmail({query: {token: this.$route.params.token as string}});
            this.success = true;
        } catch (e) {
            this.success = false;
        } finally {
            this.isLoading = false;
        }
    },
    methods: {
        goToLogin() {
            this.$router.push({ name: 'account.login' });
        },
    },
};
</script>
