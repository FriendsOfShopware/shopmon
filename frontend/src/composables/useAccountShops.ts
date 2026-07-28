import { ref } from "vue";
import { getAccountShops, type AccountShop } from "@/api/generated";

const shops = ref<AccountShop[]>([]);
const _fetched = ref(false);
let _pending: Promise<AccountShop[]> | null = null;

export function fetchAccountShops(): Promise<AccountShop[]> {
  if (_pending) {
    return _pending;
  }

  _pending = getAccountShops().then(({ data }) => {
    shops.value = data ?? [];
    _fetched.value = true;
    _pending = null;
    return shops.value;
  });

  return _pending;
}

export function resetAccountShops(): void {
  _fetched.value = false;
  _pending = null;
  shops.value = [];
}

export function useAccountShops() {
  if (!_fetched.value) {
    void fetchAccountShops();
  }
  return { shops, fetchAccountShops, resetAccountShops };
}
