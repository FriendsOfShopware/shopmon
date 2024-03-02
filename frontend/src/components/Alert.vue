<template>
    <div class="fixed top-3 right-0 flex max-w-sm overflow-hidden z-20">
        <transition
            enter-active-class="transition ease-out duration-200"
            enter-from-class="translate-x-full"
            enter-to-class="translate-x-0"
            leave-active-class="transition ease-in duration-150"
            leave-from-class="translate-x-0"
            leave-to-class="translate-x-full"
        >
            <div
                v-if="alert"
                class="relative w-screen rounded-md shadow-lg p-3 border border-l-4 border-gray-300
                flex gap-2 mr-3 mb-4 bg-white dark:bg-neutral-800 dark:border-neutral-700"
                :class="
                    { 'border-l-red-600 dark:border-l-red-600': alert.type === 'error',
                      'border-l-yellow-400 dark:border-l-yellow-400': alert.type === 'warning',
                      'border-l-green-400 dark:border-l-green-400': alert.type === 'success',
                      'border-l-sky-400 dark:border-l-sky-400': alert.type === 'info',
                    }"
            >
                <button
                    class="w-4 h-4 absolute top-1 right-1 color-red-500"
                    type="button"
                    @click="alertStore.clear()"
                >
                    <icon-fa6-solid:xmark aria-hidden="true" />
                </button>
                <div class="shrink-0">
                    <icon-fa6-solid:circle-xmark
                        v-if="alert.type === 'error'"
                        class="text-red-600 dark:text-red-400"
                    />
                    <icon-fa6-solid:circle-info
                        v-else-if="alert.type === 'warning'"
                        class="text-yellow-400 dark:text-yellow-200"
                    />
                    <icon-fa6-solid:circle-info
                        v-else-if="alert.type === 'info'"
                        class="text-sky-400"
                    />
                    <icon-fa6-solid:circle-check
                        v-else
                        class="text-green-400 dark:text-green-300"
                    />
                </div>
                <div>
                    <div class="font-medium">
                        {{ alert.title }}
                    </div>
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
