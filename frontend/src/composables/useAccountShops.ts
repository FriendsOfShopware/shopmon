import { ref } from "vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";

type AccountShop = components["schemas"]["AccountShop"];

const shops = ref<AccountShop[]>([]);
const _fetched = ref(false);
let _pending: Promise<AccountShop[]> | null = null;

export function fetchAccountShops(): Promise<AccountShop[]> {
  if (_pending) {
    return _pending;
  }

  _pending = api.GET("/account/shops").then(({ data }) => {
    shops.value = data ?? [];
    _fetched.value = true;
    _pending = null;
    return shops.value;
  });

  return _pending;
}

export function useAccountShops() {
  if (!_fetched.value) {
    fetchAccountShops();
  }
  return { shops, fetchAccountShops };
}
