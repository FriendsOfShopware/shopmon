import { ref } from "vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";

type InstanceConfig = components["schemas"]["InstanceConfig"];

const config = ref<InstanceConfig | null>(null);
let fetchPromise: Promise<void> | null = null;

export function useInstanceConfig() {
  async function load() {
    if (config.value) return;
    if (fetchPromise) {
      await fetchPromise;
      return;
    }

    fetchPromise = api
      .GET("/info/config")
      .then(({ data }) => {
        if (data) {
          config.value = data;
        }
      })
      .finally(() => {
        fetchPromise = null;
      });

    await fetchPromise;
  }

  return { config, load };
}
