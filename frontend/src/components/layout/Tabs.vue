<template>
    <div class="w-full rounded-lg bg-white mb-16 shadow overflow-hidden dark:bg-neutral-800">
        <TabGroup>
            <TabList class="border-b border-gray-200 px-6 gap-x-4 -mb-px grid dark:border-neutral-700 sm:grid-cols-2 md:grid-
                            cols-3 lg:grid-cols-none lg:auto-cols-min lg:grid-flow-col">
                <Tab v-for="(label, key) in labels" as="template" :key="key" v-slot="{ selected }">
                    <button :class="[
                    'felx items-center border-transparent hover:border-sky-500 focus:outline-none',
                    'whitespace-nowrap pt-4 pb-3 px-3 border-b-2 font-medium text-base relative',
                    'after:block after:absolute after:-bottom-[3px] after:h-px after:bg-gray-200 dark:after:bg-neutral-700 after:-left-6 after:-right-6',
                    {'border-sky-500 text-sky-600': selected},
                    ]">
                        <component v-if="label.icon" :is="label.icon" class="mr-2"></component>
                        {{ label.title }}
                        <span v-if="label.count && label.count > 0"
                            class="ml-2 bg-gray-300 rounded-full px-2.5 py-0.5 text-xs font-medium text-gray-900">
                            {{ label.count }}
                        </span>
                    </button>
                </Tab>
            </TabList>
            
            <TabPanels class="mt-px">
                <TabPanel v-for="(label, key) in labels" :key="key" class="overflow-y-auto">
                    <slot :name="`panel(${key})`" :label="label">
                        <div class="p-6">
                            <strong>{{ label.title }} Tab Panel</strong>. 
                            Use <pre class="bg-gray-200 inline-block text-xs px-1 py-0.5 rounded">&lt;template #panel({{ key }})="{ label }"&gt;...&lt;/template&gt;</pre> to fill with content
                        </div>                        
                    </slot>
                </TabPanel>
            </TabPanels>
        </TabGroup>
    </div>
</template>
  
<script setup lang="ts">
  import { ref } from 'vue'
  import { TabGroup, TabList, Tab, TabPanels, TabPanel } from '@headlessui/vue'

  defineProps<{labels: Record<string, {title: string, count?: number, icon?: string}>}>()
</script>
  