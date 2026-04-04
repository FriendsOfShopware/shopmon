<template>
  <div v-if="sortedFilteredData?.length ?? 0 > 0" class="data-table-wrapper">
    <table class="data-table">
      <thead>
        <tr>
          <th
            v-for="column in columns"
            :key="column.key"
            :class="[column.class, { sortable: column.sortable }]"
            @click="column.sortable ? setSort(column.key) : false"
          >
            {{ column.name }}
            <template v-if="column.key === sorting.sortBy">
              <icon-fa6-solid:chevron-up v-if="sorting.sortDesc" />
              <icon-fa6-solid:chevron-down v-else />
            </template>
          </th>

          <th v-if="slots['cell-actions']" class="actions">
            <slot name="cell-actions-header" />
          </th>
        </tr>
      </thead>

      <tbody>
        <tr v-for="(row, rowIndex) in sortedFilteredData" :key="rowIndex">
          <td v-for="column in columns" :key="column.key" :class="[column.class]">
            <slot :name="`cell-${column.key}`" :row="row" :row-index="rowIndex">
              {{ row[column.key] }}
            </slot>
          </td>

          <td v-if="slots['cell-actions']" class="actions">
            <slot name="cell-actions" :row="row" />
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div v-else class="data-table-empty">
    <template v-if="sortedFilteredData?.length ?? 0 === 0">
      <span>
        <template v-if="searchTerm">
          <icon-fa6-solid:circle-xmark /> {{ $t("dataTable.noSearchResult") }}
          <strong>{{ searchTerm }}</strong>
        </template>
        <template v-else> <icon-fa6-solid:circle-info /> {{ $t("dataTable.noData") }} </template>
      </span>
    </template>
    <template v-else>
      <icon-fa6-solid:inbox class="empty-icon" />
      <h2 class="empty-title">{{ $t("dataTable.noDataTitle") }}</h2>
      {{ $t("dataTable.noDataDesc") }}
    </template>
  </div>
</template>

<script
  lang="ts"
  setup
  generic="RowData extends Record<string, any>, RowKey extends Extract<keyof RowData, string>"
>
import { createNewSortInstance } from "fast-sort";
import Fuse from "fuse.js";
import { computed, useSlots, reactive } from "vue";

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
    searchTerm: "",
    defaultSort: undefined,
  },
);

const sorting = reactive({
  sortBy: props.defaultSort?.key ?? "",
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
      sensitivity: "base",
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
  const path = propertyPath.split(".");
  let value: RowData = row;

  for (const key of path) {
    value = value[key];
  }

  return value;
}

const sortedFilteredData = computed(() => {
  let data = props.data;

  if (props.searchTerm.length >= 3) {
    const fuse = new Fuse(data, {
      keys: searchableRows.value,
      threshold: 0.5,
    });

    data = fuse.search(props.searchTerm).map((result) => result.item);
  }

  if (sorting.sortBy) {
    const rowConfig = props.columns.find((row) => row.key === sorting.sortBy);

    if (!rowConfig) {
      throw new Error(`Cannot find row with key ${sorting.sortBy}, falling to index row with key`);
    }

    if (sorting.sortDesc) {
      data = naturalSort(data).desc((d) => getSortByValue(d, rowConfig.sortPath ?? sorting.sortBy));
    } else {
      data = naturalSort(data).asc((d) => getSortByValue(d, rowConfig.sortPath ?? sorting.sortBy));
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
  isolation: isolate;
  width: 100%;
  border-spacing: 0;
  text-indent: 0;
  border-color: inherit;
  border-collapse: collapse;
  background-color: var(--data-table-background);
  text-align: left;
  color: var(--text-color);

  thead {
    background-color: var(--data-table-head-background);
  }

  th {
    padding: 0.75rem;
    text-align: left;
    font-size: 1rem;
    font-weight: 600;
    color: var(--data-table-head-color);
    border-bottom: 1px solid var(--data-table-border-color);

    &.sortable {
      cursor: pointer;
    }

    .icon {
      margin-left: 0.5rem;
      font-size: 0.875rem;
      vertical-align: middle;
    }
  }

  tbody {
    td {
      padding: 0.75rem;
      background-color: var(--data-table-row-background);
      border-bottom: 1px solid var(--data-table-border-color);
    }

    tr:last-child td {
      border-bottom: 0;
    }
  }

  th,
  td {
    &.actions {
      text-align: right;

      button {
        line-height: 1.4;
      }
    }

    .icon-status {
      font-size: 1rem;
      margin-right: 0.5rem;
    }
  }
}

.data-table-empty {
  padding: 1.5rem;
  text-align: center;
  color: var(--text-color-muted);
  border: 1px solid var(--data-table-empty-border-color);
  border-radius: 0.875rem;
  background-color: var(--data-table-empty-background);

  .empty-icon {
    font-size: 9rem;
    color: var(--data-table-empty-icon-color);
    margin-bottom: 1rem;
  }

  .empty-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-color-muted);
  }
}
</style>
