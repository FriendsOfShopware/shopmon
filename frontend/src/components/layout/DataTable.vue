<template>
    <table
        v-if="sortedFilteredData?.length ?? 0 > 0"
        class="min-w-full divide-y-2 divide-gray-300 dark:divide-neutral-900"
    >
        <thead class="bg-gray-50 dark:bg-neutral-800">
            <tr>
                <th
                    v-if="slots['cell-favicon']"
                    class="py-3.5 px-3 text-left whitespace-nowrap"
                />
                <th
                    v-for="(column, index) in columns"
                    :key="column.key"
                    class="py-3.5"
                    :class="[{
                        'px-3 text-left whitespace-nowrap': !column.class && !column.classOverride,
                        'pl-4 lg:pl-8': index === 0 && !column.classOverride,
                        'pr-4 sm:pr-6 lg:pr-8': index === Object.keys(columns).length - 1 && !column.classOverride,
                        'cursor-pointer': column.sortable,
                    }, column.class]"
                    @click="column.sortable ? setSort(column.key) : false"
                >
                    {{ column.name }}

                    <template v-if="column.key === sorting.sortBy">
                        <icon-fa6-solid:chevron-up v-if="sorting.sortDesc" />
                        <icon-fa6-solid:chevron-down v-else />
                    </template>
                    <span
                        v-else
                        class="inline-block w-[14px]"
                    />
                </th>
                <th
                    v-if="slots['cell-actions']"
                    class="py-3.5 px-3 text-left whitespace-nowrap"
                />
            </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 dark:divide-none">
            <tr
                v-for="(row, rowIndex) in sortedFilteredData"
                :key="rowIndex"
                class="group even:bg-gray-50 hover:bg-sky-50 dark:even:bg-[#2b2b2b] dark:hover:bg-[#2a2b2f]"
            >
                <td v-if="slots['cell-favicon']">
                    <slot
                        name="cell-favicon"
                        :row="row"
                    />
                </td>
                <td
                    v-for="column in columns"
                    :key="column.key"
                    class="whitespace-nowrap py-4 align-middle"
                    :class="[{
                        'px-3': !column.class && !column.classOverride,
                        'pl-4 lg:pl-8': rowIndex === 0 && !column.classOverride,
                        'pr-4 sm:pr-6 lg:pr-8': rowIndex === Object.keys(columns).length - 1 && !column.classOverride
                    }, column.class ]"
                >
                    <slot
                        :name="`cell-${column.key}`"
                        :row="row"
                    >
                        {{ row[column.key] }}
                    </slot>
                </td>

                <td
                    v-if="slots['cell-actions']"
                    class="px-3 text-right"
                >
                    <slot
                        name="cell-actions"
                        :row="row"
                    />
                </td>
            </tr>
        </tbody>
    </table>
    <div
        v-else
        class="p-6 text-center text-gray-400"
    >
        <template v-if="sortedFilteredData?.length ?? 0 === 0">
            <span class="text-lg">
                <template v-if="searchTerm">
                    <icon-fa6-solid:circle-xmark /> no search result for <strong>{{ searchTerm }}</strong>
                </template>
                <template v-else>
                    <icon-fa6-solid:circle-info /> no data
                </template>
            </span>
        </template>
        <template v-else>
            <icon-fa6-solid:inbox class="text-9xl text-gray-200 mb-4" />
            <h2 class="text-2xl font-bold text-gray-400 block">
                No Data
            </h2>
            There is no data to display the table
        </template>
    </div>
</template>

<script lang="ts" setup generic="RowData extends Record<string, any >, RowKey extends Extract<keyof RowData, string>">
import { computed, useSlots } from 'vue';
import { createNewSortInstance } from 'fast-sort';
import Fuse from 'fuse.js';

const slots = useSlots();

const props = defineProps<{
    columns: Array<{
        key: RowKey,
        name: string,
        class?: string,
        classOverride?: boolean,
        searchable?: boolean,
        sortable?: boolean,
        sortPath?: string,
    }>,
    searchTerm?: string,
    sorting: {
        sortBy: RowKey,
        sortDesc: boolean,
    },
    data: RowData[]
}>();

props.columns.at;
//     ^?
const emits = defineEmits<{
    'update:searchTerm': [value: string],
    'update:sorting': [value: {
        sortBy: string,
        sortDesc: boolean,
    }],
}>();

const naturalSort = createNewSortInstance({
    comparer: (a, b, order) => {
        // Always sort null values at last
        if (a === null || a === undefined) {
            return order;
        } else if (b === null || b === undefined) {
            return (order * -1);
        }

        const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' });
        return collator.compare(a, b);
    },
});

function setSort(key: string) {
    const sorting = structuredClone(props.sorting);

    if (sorting.sortBy !== key) {
        emits('update:sorting',{
            sortBy: key,
            sortDesc: false,
        });

        return;
    }

    emits('update:sorting', {
        sortBy: key,
        sortDesc: !sorting.sortDesc,
    });
}

const searchableRows = computed(() => {
    return props.columns.filter(row => row.searchable).map(row => row.key);
});

function getSortByValue(row: RowData, propertyPath: string): unknown {
    const path = propertyPath.split('.');
    let value = row;

    for (const key of path) {
        value = value[key] as unknown;
    }

    return value;
}

const sortedFilteredData = computed(() => {
    let data = structuredClone(props.data);
    const sortBy = props.sorting.sortBy;
    const sortDesc = props.sorting.sortDesc;

    if (props.searchTerm && props.searchTerm.length >= 3) {
        const fuse = new Fuse(data, {
            keys: searchableRows.value,
            threshold: 0,
        });

        data = fuse.search(props.searchTerm).map(result => result.item);
    }

    if (sortBy) {
        const rowConfig = props.columns.find(row => row.key === sortBy);

        if (!rowConfig) {
            throw new Error(`Cannot find row with key ${sortBy}, falling to index row with key`);
        }

        if(sortDesc) {
            data = naturalSort(data).desc(d => getSortByValue(d, rowConfig.sortPath ?? sortBy));
        } else {
            data = naturalSort(data).desc(d => getSortByValue(d, rowConfig.sortPath ?? sortBy));
        }
    }

    return data;
});
</script>
