<script setup>
import { InformationCircleIcon, CheckCircleIcon  } from '@heroicons/vue/solid'

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
    <div class="rounded-md bg-blue-50 p-4" v-if="isLoading">
        <div class="flex">
            <div class="flex-shrink-0">
                <InformationCircleIcon class="h-5 w-5 text-blue-400" aria-hidden="true" />
            </div>
            <div class="ml-3 flex-1 md:flex md:justify-between">
                <p class="text-sm text-blue-700">Confirming your Mail Address</p>
            </div>
        </div>
    </div>

    <div class="rounded-md bg-green-50 p-4" v-if="!isLoading && success">
        <div class="flex">
            <div class="flex-shrink-0">
                <CheckCircleIcon class="h-5 w-5 text-green-400" aria-hidden="true" />
            </div>
            <div class="ml-3">
                <h3 class="text-sm font-medium text-green-800">Mail Confirmation</h3>
                <div class="mt-2 text-sm text-green-700">
                    <p>Your Mail address has been verified</p>
                </div>
                <div class="mt-4">
                    <div class="-mx-2 -my-1.5 flex">
                        <button type="button"
                            @click="goToLogin"
                            class="bg-green-50 px-2 py-1.5 rounded-md text-sm font-medium text-green-800 hover:bg-green-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-green-50 focus:ring-green-600">Go to login</button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
