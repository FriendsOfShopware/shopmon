import { ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

export type Defaults = {
  search?: string;
  page?: number;
  filters: Record<string, string>;
};

function parseQuery(
  query: Record<string, unknown>,
  defaultSearch: string,
  defaultPage: number,
  defaultFilters: Record<string, string>,
) {
  const qVal = query.q;
  const search = typeof qVal === "string" ? qVal : defaultSearch;

  const pageVal = query.page;
  let page = defaultPage;
  if (typeof pageVal === "string") {
    const parsed = parseInt(pageVal, 10);
    page = isNaN(parsed) ? defaultPage : Math.max(1, parsed);
  }

  const filters: Record<string, string> = {};
  for (const key of Object.keys(defaultFilters)) {
    const routeVal = query[key];
    if (typeof routeVal === "string" && routeVal !== "") {
      filters[key] = routeVal;
    } else {
      filters[key] = defaultFilters[key];
    }
  }

  return { search, page, filters };
}

/** Stable key for comparing list query state (ignores unrelated route query keys). */
function stateKey(
  search: string,
  page: number,
  filters: Record<string, string>,
  defaultSearch: string,
  defaultPage: number,
  defaultFilters: Record<string, string>,
) {
  const parts: string[] = [];
  if (search && search !== defaultSearch) parts.push(`q=${search}`);
  if (page && page !== defaultPage) parts.push(`page=${page}`);
  for (const key of Object.keys(defaultFilters).sort()) {
    const val = filters[key];
    if (val !== undefined && val !== defaultFilters[key]) {
      parts.push(`${key}=${val}`);
    }
  }
  return parts.join("&");
}

export function useAdminListQuery(defaults: Defaults) {
  const route = useRoute();
  const router = useRouter();

  const defaultSearch = defaults.search ?? "";
  const defaultPage = defaults.page ?? 1;
  const defaultFilters = { ...defaults.filters };

  const initial = parseQuery(route.query, defaultSearch, defaultPage, defaultFilters);

  const search = ref(initial.search);
  const page = ref(initial.page);
  const filters = ref<Record<string, string>>(initial.filters);

  /**
   * Bumped when the route query changes list state from outside this composable
   * (e.g. browser back/forward or same-route navigation). List views should
   * watch this and reload. Own syncToUrl updates do not bump it.
   */
  const revision = ref(0);

  function currentKey() {
    return stateKey(
      search.value,
      page.value,
      filters.value,
      defaultSearch,
      defaultPage,
      defaultFilters,
    );
  }

  function applyFromRoute() {
    const next = parseQuery(route.query, defaultSearch, defaultPage, defaultFilters);
    const nextKey = stateKey(
      next.search,
      next.page,
      next.filters,
      defaultSearch,
      defaultPage,
      defaultFilters,
    );
    if (currentKey() === nextKey) return false;

    search.value = next.search;
    page.value = next.page;
    filters.value = next.filters;
    return true;
  }

  watch(
    () => route.query,
    () => {
      if (applyFromRoute()) {
        revision.value++;
      }
    },
  );

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
    revision,
    syncToUrl,
    reset,
  };
}
