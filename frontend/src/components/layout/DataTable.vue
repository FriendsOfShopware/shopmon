<template>
  <div v-if="sortedFilteredData?.length ?? 0 > 0" class="overflow-x-auto">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead
            v-for="column in columns"
            :key="column.key"
            :class="[column.class, { 'cursor-pointer select-none': column.sortable }]"
            @click="column.sortable ? setSort(column.key) : false"
          >
            {{ column.name }}
            <template v-if="column.key === sorting.sortBy">
              <icon-fa6-solid:chevron-up v-if="sorting.sortDesc" class="ml-1 inline size-3" />
              <icon-fa6-solid:chevron-down v-else class="ml-1 inline size-3" />
            </template>
          </TableHead>

          <TableHead v-if="slots['cell-actions']" class="text-right">
            <slot name="cell-actions-header" />
          </TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        <TableRow v-for="(row, rowIndex) in sortedFilteredData" :key="rowIndex">
          <TableCell v-for="column in columns" :key="column.key" :class="[column.class]">
            <slot :name="`cell-${column.key}`" :row="row" :row-index="rowIndex">
              {{ row[column.key] }}
            </slot>
          </TableCell>

          <TableCell v-if="slots['cell-actions']" class="text-right">
            <slot name="cell-actions" :row="row" />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>

  <div
    v-else
    class="flex flex-col items-center gap-2 rounded-xl border bg-card p-6 text-center text-muted-foreground"
  >
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
      <icon-fa6-solid:inbox class="size-36 text-border" />
      <h2 class="text-2xl font-bold text-muted-foreground">{{ $t("dataTable.noDataTitle") }}</h2>
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

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
