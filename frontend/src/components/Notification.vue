<template>
    <div class="notifications-container">
        <transition 
            enter-active-class="transition ease-out duration-200" 
            enter-from-class="translate-x-full"
            enter-to-class="translate-x-0" 
            leave-active-class="transition ease-in duration-150"
            leave-from-class="translate-x-0" 
            leave-to-class="translate-x-full"
        >
            <div v-if="alert" class="notification" :class="`notification-${alert.type}`">
                <button class="notification-close" type="button" @click="alertStore.clear()">
                    <icon-fa6-solid:xmark aria-hidden="true" />
                </button>
                <div class="notification-icon">
                    <icon-fa6-solid:circle-xmark v-if="alert.type === 'error'" class="icon-error" />
                    <icon-fa6-solid:circle-info v-else-if="alert.type === 'warning'" class="icon-warning" />
                    <icon-fa6-solid:circle-info v-else-if="alert.type === 'info'" class="icon-info" />
                    <icon-fa6-solid:circle-check v-else class="icon-success" />
                </div>
                <div class="notification-content">
                    <div class="notification-title">{{ alert.title }}</div>
                    {{ alert.message }}
                </div>
            </div>
        </transition>
    </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAlertStore } from '@/stores/alert.store';

const alertStore = useAlertStore();
const { alert } = storeToRefs(alertStore);
</script>

<style scoped>
.notifications-container {
    position: fixed;
    top: 0.75rem;
    right: 0;
    z-index: 20;
    display: flex;
    max-width: 24rem;
    overflow: hidden;
}

.notification {
    position: relative;
    width: 100vw;
    padding: .75rem;
    border-radius: 0.375rem;
    border: 1px solid var(--notification-border-color);
    border-left-width: 5px;
    background-color: var(--panel-background);
    margin: 0 0.75rem 1rem 0;
    display: flex;
    gap: 0.5rem;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);

    &-info {
        border-left-color: var(--info-color);
    }

    &-success {
        border-left-color: var(--success-color);
    }

    &-warning {
        border-left-color: var(--warning-color);
    }

    &-error {
        border-left-color: var(--error-color);
    }

    &-close {
        position: absolute;
        top: 0.25rem;
        right: 0.25rem;
        width: 1rem;
        height: 1rem;
        display: flex;
        align-items: center;
        justify-content: center;
        opacity: .5;

        &:hover {
            opacity: 1;
        }
    }

    &-icon {
        display: flex;
        justify-content: center;
        margin-top: .25rem;
    }

    &-content {
        flex: 1;
    }

    &-title {
        font-weight: 500;
    }

    .dark & {
        box-shadow: none;
    }
}
</style>
