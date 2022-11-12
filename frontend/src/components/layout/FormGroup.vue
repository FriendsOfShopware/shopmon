<script lang="ts" setup>
    defineProps<{ title: string, subTitle?: string, info?: string }>()
</script>

<template>
    <div class="md:grid md:grid-cols-3 md:gap-6 pb-12">
        <div class="md:col-span-1">
            <h3 class="text-lg font-medium leading-6">
                {{ title }}
            </h3>
            <p v-if="subTitle" class="text-gray-600">
                {{ subTitle }}
            </p>

            <div v-if="info" class="rounded-md bg-blue-0 p-4 border border-blue-300 flex mt-3 dark:bg-gray-900 dark:border-sky-400">
                <div class="flex-shrink-0">
                    <icon-fa6-solid:circle-info class="h-5 w-5 text-sky-400" aria-hidden="true" />
                </div>
                <div class="ml-3 flex-1 md:flex md:justify-between text-sky-900 dark:text-sky-600" v-html="info">
                </div>
            </div>
        </div>

        <div class="mt-5 md:mt-0 md:col-span-2">
            <div class="px-4 py-5 bg-white space-y-6 sm:p-6 shadow overflow-hidden rounded-md dark:bg-neutral-800 dark:shadow-none">
                <slot />
            </div>
        </div>
    </div>
</template>