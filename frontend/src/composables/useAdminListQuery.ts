import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";

export type Defaults = {
  search?: string;
  page?: number;
  filters: Record<string, string>;
};

export function useAdminListQuery(defaults: Defaults) {
  const route = useRoute();
  const router = useRouter();

  const defaultSearch = defaults.search ?? "";
  const defaultPage = defaults.page ?? 1;
  const defaultFilters = { ...defaults.filters };

  // Read initial route values
  const qVal = route.query.q;
  const initialSearch = typeof qVal === "string" ? qVal : defaultSearch;

  const pageVal = route.query.page;
  let initialPage = defaultPage;
  if (typeof pageVal === "string") {
    const parsed = parseInt(pageVal, 10);
    initialPage = isNaN(parsed) ? defaultPage : Math.max(1, parsed);
  }

  const initialFilters: Record<string, string> = {};
  for (const key of Object.keys(defaultFilters)) {
    const routeVal = route.query[key];
    if (typeof routeVal === "string" && routeVal !== "") {
      initialFilters[key] = routeVal;
    } else {
      initialFilters[key] = defaultFilters[key];
    }
  }

  const search = ref(initialSearch);
  const page = ref(initialPage);
  const filters = ref<Record<string, string>>(initialFilters);

  function syncToUrl() {
    const query = { ...route.query };

    // Update search (q)
    if (search.value && search.value !== defaultSearch) {
      query.q = search.value;
    } else {
      delete query.q;
    }

    // Update page
    if (page.value && page.value !== defaultPage) {
      query.page = String(page.value);
    } else {
      delete query.page;
    }

    // Update filters
    for (const key of Object.keys(defaultFilters)) {
      const val = filters.value[key];
      if (val !== undefined && val !== defaultFilters[key]) {
        query[key] = val;
      } else {
        delete query[key];
      }
    }

    void router.replace({ query });
  }

  function reset() {
    search.value = defaultSearch;
    page.value = defaultPage;
    filters.value = { ...defaultFilters };
    syncToUrl();
  }

  return {
    search,
    page,
    filters,
    syncToUrl,
    reset,
  };
}
