<template>
    <div
        v-if="sortedFilteredData?.length ?? 0 > 0"
        class="data-table-wrapper"
    >
        <table class="data-table">
            <thead>
            <tr>
                <th
                    v-for="column in columns"
                    :key="column.key"
                    :class="[
                        column.class,
                        { sortable: 'sortable' },
                    ]"
                    @click="column.sortable ? setSort(column.key) : false"
                >
                    {{ column.name }}
                    <template v-if="column.key === sorting.sortBy">
                        <icon-fa6-solid:chevron-up v-if="sorting.sortDesc" />
                        <icon-fa6-solid:chevron-down v-else />
                    </template>
                </th>

                <th
                    v-if="slots['cell-actions']"
                    class="actions"
                >
                    <slot name="cell-actions-header" />
                </th>
            </tr>
            </thead>

            <tbody>
            <tr
                v-for="(row, rowIndex) in sortedFilteredData"
                :key="rowIndex"
            >
                <td
                    v-for="column in columns"
                    :key="column.key"
                    :class="[column.class]"
                >
                    <slot
                        :name="`cell-${column.key}`"
                        :row="row"
                        :row-index="rowIndex"
                    >
                        {{ row[column.key] }}
                    </slot>
                </td>

                <td
                    v-if="slots['cell-actions']"
                    class="actions"
                >
                    <slot
                        name="cell-actions"
                        :row="row"
                    />
                </td>
            </tr>
            </tbody>
        </table>
    </div>


    <div
        v-else
        class="data-table-empty"
    >
        <template v-if="sortedFilteredData?.length ?? 0 === 0">
            <span>
                <template v-if="searchTerm">
                    <icon-fa6-solid:circle-xmark /> no search result for <strong>{{ searchTerm }}</strong>
                </template>
                <template v-else>
                    <icon-fa6-solid:circle-info /> no data
                </template>
            </span>
        </template>
        <template v-else>
            <icon-fa6-solid:inbox class="empty-icon" />
            <h2 class="empty-title">
                No Data
            </h2>
            There is no data to display the table
        </template>
    </div>
</template>

<script lang="ts" setup generic="RowData extends Record<string, any>, RowKey extends Extract<keyof RowData, string>">
import { createNewSortInstance } from 'fast-sort';
import Fuse from 'fuse.js';
import { computed, useSlots } from 'vue';
import { reactive } from 'vue';

const slots = useSlots();

const props = withDefaults(
    defineProps<{
        columns: Array<{
            key: RowKey;
            name: string;
            class?: string;
            searchable?: boolean;
            sortable?: boolean;
            sortPath?: string;
        }>;
        defaultSort?: {
            key: RowKey;
            desc?: boolean;
        };
        data: RowData[];
        searchTerm?: string;
    }>(),
    {
        searchTerm: '',
        defaultSort: undefined,
    },
);

const sorting = reactive({
    sortBy: props.defaultSort?.key ?? '',
    sortDesc: props.defaultSort?.desc ?? false,
});

const naturalSort = createNewSortInstance({
    comparer: (a, b, order) => {
        // Always sort null values at last
        if (a === null || a === undefined) {
            return order;
        }
        if (b === null || b === undefined) {
            return order * -1;
        }

        const collator = new Intl.Collator(undefined, {
            numeric: true,
            sensitivity: 'base',
        });
        return collator.compare(a, b);
    },
});

function setSort(key: string) {
    if (sorting.sortBy !== key) {
        sorting.sortBy = key;
        sorting.sortDesc = false;
        return;
    }

    sorting.sortDesc = !sorting.sortDesc;
}

const searchableRows = computed(() => {
    return props.columns.filter((row) => row.searchable).map((row) => row.key);
});

function getSortByValue(row: RowData, propertyPath: string): unknown {
    const path = propertyPath.split('.');
    let value: RowData = row;

    for (const key of path) {
        value = value[key];
    }

    return value;
}

const sortedFilteredData = computed(() => {
    let data = props.data;
    const sortBy = sorting.sortBy;
    const sortDesc = sorting.sortDesc;

    if (props.searchTerm.length >= 3) {
        const fuse = new Fuse(data, {
            keys: searchableRows.value,
            threshold: 0.5,
        });

        data = fuse.search(props.searchTerm).map((result) => result.item);
    }

    if (sortBy) {
        const rowConfig = props.columns.find((row) => row.key === sortBy);

        if (!rowConfig) {
            throw new Error(
                `Cannot find row with key ${sortBy}, falling to index row with key`,
            );
        }

        if (sortDesc) {
            data = naturalSort(data).desc((d) =>
                getSortByValue(d, rowConfig.sortPath ?? sortBy),
            );
        } else {
            data = naturalSort(data).asc((d) =>
                getSortByValue(d, rowConfig.sortPath ?? sortBy),
            );
        }
    }

    return data;
});
</script>

<style>
.data-table-wrapper {
    overflow-x: auto;
}

.data-table {
    width: 100%;
    border-spacing: 0;
    text-indent: 0;
    border-color: inherit;
    border-collapse: collapse;

    thead {
        background-color: var(--data-table-head-background);
    }

    th {
        padding: 1rem 0.75rem;
        text-align: left;
        font-weight: 500;
        color: var(--data-table-head-color);
        white-space: nowrap;

        &.sortable {
            cursor: pointer;
        }
    }

    tbody {
        border-top: 2px solid var(--background-color);

        tr {
            background-color: var(--data-table-row-background);

            &:nth-child(odd) {
                background-color: var(--data-table-row-odd-background);
            }

            &:hover {
                background-color: var(--data-table-row-hover-background);
            }
        }

        td {
            padding: 1rem 0.75rem;
        }
    }

    th, td {
        &:first-child {
            padding-left: 1rem;

            @media (min-width: 1024px) {
                padding-left: 1.5rem;
            }
        }

        &.actions {
            text-align: right;
            padding-right: 1rem;

            @media (min-width: 640px) {
                padding-right: 1.5rem;
            }

            @media (min-width: 1024px) {
                padding-right: 2rem;
            }

            button {
                line-height: 1.4;
            }

            .icon {
                opacity: .5;

                &:hover {
                    opacity: 1;
                }
            }
        }

        .icon-status {
            font-size: 1rem;
            margin-right: .5rem;
        }
    }
}

.data-table-empty {
    padding: 1.5rem;
    text-align: center;
    color: #9ca3af;

    .empty-icon {
        font-size: 9rem;
        color: #e5e7eb;
        margin-bottom: 1rem;
    }

    .empty-title {
        font-size: 1.5rem;
        font-weight: 700;
        color: #9ca3af;
    }
}
</style>
