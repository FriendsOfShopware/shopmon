<script lang="ts" setup>
    defineProps<{labels: Object, data: Object}>()
</script>
    
<template>
    <table class="min-w-full divide-y-2 divide-gray-300 background">
        <thead class="bg-gray-50">
            <tr>
                <th v-for="(label, key, index) in labels" 
                    :key="label" 
                    class="py-3.5"
                    :class="[
                        {'px-3 text-left': !label.class},
                        {'pl-4 lg:pl-8': index === 0},
                        {'pr-4 sm:pr-6 lg:pr-8': index === Object.keys(labels).length - 1},
                        label.class,
                    ]"
                >
                    {{ label.name }}
                </th>
            </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
            <tr class="even:bg-gray-50" v-for="(item, key) in data" :key="key">
                <td v-for="(label, key, index) in labels"
                    class="whitespace-nowrap py-4 align-middle"
                    :class="[
                        {'px-3': !label.class},
                        {'pl-4 lg:pl-8': index === 0},
                        {'pr-4 sm:pr-6 lg:pr-8': index === Object.keys(labels).length - 1},
                        label.class,
                    ]"
                >
                    <slot :name="`cell(${key})`" :value="item[key]" :item="item">
                        {{ item[key] }}
                    </slot>
                </td>
            </tr>
        </tbody>
    </table>
</template>