<template>
    <div v-if="isImpersonating" class="impersonation-banner">
        <div class="impersonation-content">
            <div class="impersonation-text">
                <icon-fa6-solid:user-secret class="impersonation-icon" />
                <span>You are currently impersonating <strong>{{ session?.data?.user?.email }}</strong></span>
            </div>
            <button class="btn btn-sm btn-stop" @click="stopImpersonating">
                Stop Impersonating
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { authClient } from '@/helpers/auth-client';
import { computed } from 'vue';

const session = authClient.useSession();
const alert = useAlert();

const isImpersonating = computed(() => {
    // Check if the session has impersonation data
    const sessionData = session.value?.data;
    if (!sessionData) return false;

    // Better-auth stores impersonatedBy in the session object
    return !!sessionData.session?.impersonatedBy;
});

async function stopImpersonating() {
    try {
        await authClient.admin.stopImpersonating();
        // Force a complete page reload to ensure clean session state
        window.location.reload();
    } catch (error) {
        alert.error(
            `Failed to stop impersonating.  Error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        );
    }
}
</script>

<style scoped>
.impersonation-banner {
    background-color: #f59e0b;
    color: #ffffff;
    padding: 0.75rem 0;
    position: sticky;
    top: 0;
    z-index: 10;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.impersonation-content {
    max-width: 1280px;
    margin: 0 auto;
    padding: 0 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
}

.impersonation-text {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
}

.impersonation-icon {
    width: 1rem;
    height: 1rem;
}

.btn-stop {
    background-color: rgba(255, 255, 255, 0.2);
    color: #ffffff;
    border: 1px solid rgba(255, 255, 255, 0.3);
    padding: 0.25rem 0.75rem;
    font-size: 0.875rem;
}

.btn-stop:hover {
    background-color: rgba(255, 255, 255, 0.3);
    border-color: rgba(255, 255, 255, 0.4);
}

@media (max-width: 640px) {
    .impersonation-content {
        flex-direction: column;
        text-align: center;
    }
}
</style>
