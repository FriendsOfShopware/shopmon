import { ref } from "vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";

type AccountEnvironment = components["schemas"]["AccountEnvironment"];

const environments = ref<AccountEnvironment[]>([]);
const _fetched = ref(false);
let _pending: Promise<AccountEnvironment[]> | null = null;

export function fetchAccountEnvironments(): Promise<AccountEnvironment[]> {
  if (_pending) {
    return _pending;
  }

  _pending = api.GET("/account/environments").then(({ data }) => {
    environments.value = data ?? [];
    _fetched.value = true;
    _pending = null;
    return environments.value;
  });

  return _pending;
}

export function resetAccountEnvironments(): void {
  _fetched.value = false;
  _pending = null;
  environments.value = [];
}

export function useAccountEnvironments() {
  if (!_fetched.value) {
    fetchAccountEnvironments();
  }
  return { environments, fetchAccountEnvironments, resetAccountEnvironments };
}
