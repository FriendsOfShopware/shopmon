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
            <router-link to="/account/login">
                Login
            </router-link>
        </Alert>

        <Alert type="error" v-else>
            The given token has been expired
        </Alert>
    </template>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store';
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
        const authStore = useAuthStore();

        try {
            await authStore.confirmMail(this.$route.params.token as string);
            this.success = true;
        } catch (e) {
            this.success = false;
        } finally {
            this.isLoading = false;
        }
    },
    methods: {
        goToLogin() {
            this.$router.push('/account/login');
        },
    },
};
</script>
