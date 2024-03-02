<template>
    <div class="w-full rounded-lg bg-white mb-16 shadow overflow-hidden dark:bg-neutral-800">
        <tab-group>
            <tab-list
                class="border-b border-gray-200 px-6 gap-x-4 -mb-px grid dark:border-neutral-700 sm:grid-cols-2 md:grid-
                            cols-3 lg:grid-cols-none lg:auto-cols-min lg:grid-flow-col"
            >
                <tab
                    v-for="label in props.labels"
                    :key="label.key"
                    v-slot="{ selected }"
                    as="template"
                >
                    <button
                        class="flex items-center focus:outline-none whitespace-nowrap pt-4 pb-3 px-3 font-medium text-base
                        relative before:block before:z-1 before:w-0 before:absolute before:left-1/2 before:-translate-x-1/2
                        before:-bottom-[1px] before:h-[3px] before:bg-sky-500 before:hover:w-full before:transition-all
                        after:block after:absolute after:-bottom-[1px] after:h-px after:bg-gray-200 dark:after:bg-neutral-700
                        after:-left-6 after:-right-6"
                        :class="{
                            'border-sky-500 text-sky-600 before:!w-full': selected
                        }"
                        type="button"
                    >
                        <component
                            :is="label.icon"
                            v-if="label.icon"
                            class="mr-2"
                        />
                        {{ label.title }}
                        <span
                            v-if="label.count !== undefined"
                            class="ml-2 bg-gray-300 rounded-full px-2.5 py-0.5 text-xs font-medium text-gray-900"
                        >
                            {{ label.count }}
                        </span>
                    </button>
                </tab>
            </tab-list>

            <tab-panels class="mt-px">
                <tab-panel
                    v-for="label in props.labels"
                    :key="label.key"
                    class="overflow-y-auto"
                >
                    <slot
                        :name="`panel-${label.key}`"
                        :label="label"
                    >
                        <div class="p-6">
                            <strong>{{ label.title }} Tab Panel</strong>.
                            Use
                            <pre
                                class="bg-gray-200 inline-block text-xs px-1 py-0.5 rounded"
                            >&lt;template #panel-{{ label.key }}="{ label }"&gt;...&lt;/template&gt;</pre>
                            to fill with content
                        </div>
                    </slot>
                </tab-panel>
            </tab-panels>
        </tab-group>
    </div>
</template>

<script setup lang="ts" generic="T extends string">
import { TabGroup, TabList, Tab, TabPanels, TabPanel } from '@headlessui/vue';
import type { FunctionalComponent } from 'vue';

const props = defineProps<{
    labels: Array<{
        key: string,
        title: string,
        count?: number,
        icon?: FunctionalComponent
    }>
}>();
</script>
