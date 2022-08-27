<script setup>
import { InformationCircleIcon, CheckCircleIcon  } from '@heroicons/vue/24/solid'

import { useAuthStore } from '@/stores';
</script>

<script>
export default {
    data() {
        return {
            isLoading: true,
            success: false,
        }
    },
    async created() {
        const authStore = useAuthStore();

        try {
            await authStore.confirmMail(this.$route.params.token);
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
    }
}
</script>

<template>
    <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">Confirming your Account Registration</h4>
    <div class="rounded-md bg-blue-50 p-4" v-if="isLoading">
        <div class="flex">
            <div class="flex-shrink-0">
                <InformationCircleIcon class="h-5 w-5 text-blue-400" aria-hidden="true" />
            </div>
            <div class="ml-3 flex-1 md:flex md:justify-between">
                <p class="text-sm text-blue-700">Loading...</p>
            </div>
        </div>
    </div>

    <div v-else>
        <div class="rounded-md bg-green-50 p-4" v-if="success">
            <div class="flex">
                <div class="flex-shrink-0">
                    <CheckCircleIcon class="h-5 w-5 text-blue-400" aria-hidden="true" />
                </div>
                <div class="ml-3 flex-1 md:flex md:justify-between">
                    <p class="text-sm text-blue-700">Your Mail Address has been confirmed. 
                        <router-link to="/account/login" class="text-blue-700">
                            Login
                        </router-link>
                    </p>
                </div>
            </div>
        </div>

        <div class="rounded-md bg-red-50 p-4" v-else>
            <div class="flex">
                <div class="flex-shrink-0">
                    <InformationCircleIcon class="h-5 w-5 text-red-400" aria-hidden="true" />
                </div>
                <div class="ml-3 flex-1 md:flex md:justify-between">
                    <p class="text-sm text-red-700">The given token has been expired</p>
                </div>
            </div>
        </div>
    </div>
</template>
