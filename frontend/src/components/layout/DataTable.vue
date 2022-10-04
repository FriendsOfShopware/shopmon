 <template>
    <table class="min-w-full divide-y-2 divide-gray-300 dark:divide-neutral-900" v-if="data.length > 0">
        <thead class="bg-gray-50 dark:bg-neutral-800">
            <tr>
                <th v-for="(label, key, index) in labels" 
                    :key="key" 
                    class="py-3.5"
                    :class="[
                        {'px-3 text-left': !label.class && !label.classOverride},
                        {'pl-4 lg:pl-8': index === 0 && !label.classOverride},
                        {'pr-4 sm:pr-6 lg:pr-8': index === Object.keys(labels).length - 1 && !label.classOverride},
                        {'cursor-pointer': label.sortable},
                        label.class,
                    ]"
                    @click="label.sortable ? setSort(key) : false"
                >
                    {{ label.name }}

                    <template v-if="sortBy === key">
                        <icon-fa6-solid:chevron-up v-if="sortDesc === true" />
                        <icon-fa6-solid:chevron-down v-else />
                    </template>
                    <span class="inline-block w-[14px]" v-else></span>
                </th>
            </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 dark:divide-none">
            <tr class="group even:bg-gray-50 hover:bg-sky-50 dark:even:bg-[#2b2b2b] dark:hover:bg-[#2a2b2f]" v-for="(item, itemKey) in sortedData" :key="itemKey">
                <td v-for="(label, key, index) in labels"
                    class="whitespace-nowrap py-4 align-middle"
                    :class="[
                        {'px-3': !label.class && !label.classOverride},
                        {'pl-4 lg:pl-8': index === 0 && !label.classOverride},
                        {'pr-4 sm:pr-6 lg:pr-8': index === Object.keys(labels).length - 1 && !label.classOverride},
                        label.class,
                    ]"
                    :key="key"
                >
                    <slot :name="`cell(${key})`" :value="item[key]" :item="item" :data="sortedData" :itemKey="itemKey">
                        {{ item[key] }}
                    </slot>
                </td>
            </tr>
        </tbody>
    </table>
    <div class="p-6 text-center text-gray-400" v-else>
        <icon-fa6-solid:inbox class="text-9xl text-gray-200 mb-4" />
        <h2 class="text-2xl font-bold text-gray-400 block">No Data</h2>
        There is no data to display the table
    </div>
</template>

<script lang="ts" setup>
    import { computed, reactive, watch, ref } from 'vue';
    import { sort, createNewSortInstance } from 'fast-sort';

    const props = defineProps<{labels: Record<string, {class?: string, classOverride?: boolean, name: string, sortable?: boolean}>, defaultSorting?: {by: string, desc?: boolean},data: Record<string, any>[]}>()

    const sortBy = ref(props.defaultSorting?.by || null);
    const sortDesc = ref(props.defaultSorting?.desc || null);

    const naturalSort = createNewSortInstance({
        comparer: new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' }).compare,
    });

    const sortedData = computed(() => {
        const { data } = props;
        if (sortBy.value === null) return data;

        if (sortDesc.value) {
            return naturalSort(data).desc(sortBy.value);
        } else {
            return naturalSort(data).asc(sortBy.value);
        }
    });

    const setSort = (key: string) => {
        if (sortBy.value === key) {
            if (sortDesc.value === null) sortDesc.value = true;
            else {
                sortBy.value = null;
                sortDesc.value = null;
            }
        } else {
            sortBy.value = key;
            sortDesc.value = null;
        }
    };
</script>