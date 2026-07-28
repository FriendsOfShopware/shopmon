import { ref } from "vue";
import { getInstanceConfig, type InstanceConfig } from "@/api/generated";

const config = ref<InstanceConfig | null>(null);
let fetchPromise: Promise<void> | null = null;

export function useInstanceConfig() {
  async function load() {
    if (config.value) return;
    if (fetchPromise) {
      await fetchPromise;
      return;
    }

    fetchPromise = getInstanceConfig()
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
