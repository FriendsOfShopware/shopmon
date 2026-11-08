<template>
  <section class="space-y-3 border-t pt-4">
    <div>
      <h3 class="text-sm font-semibold">{{ $t("destinations.sharedSettings") }}</h3>
      <p class="text-xs text-muted-foreground">{{ $t("destinations.sharedSettingsHint") }}</p>
    </div>
    <p v-if="loading" role="status" class="text-sm text-muted-foreground">
      {{ $t("destinations.loading") }}
    </p>
    <div v-else-if="failed" role="alert" class="text-sm text-destructive">
      {{ $t("destinations.failedLoad") }}
      <button type="button" class="underline" @click="load">{{ $t("destinations.retry") }}</button>
    </div>
    <template v-else-if="destinations.length">
      <label class="block space-y-1 text-sm"
        ><span>{{ $t("destinations.destination") }}</span>
        <select
          v-model="selectedId"
          class="h-9 w-full rounded-md border bg-background px-3 text-sm"
        >
          <option v-for="destination in destinations" :key="destination.id" :value="destination.id">
            {{ destination.name }} ({{ destination.provider }})
          </option>
        </select>
      </label>
      <p class="text-xs text-muted-foreground">
        {{ $t("destinations.environmentPreferenceHint") }}
      </p>
      <NotificationMatrix
        v-if="selectedId"
        :key="`${environmentId}:${selectedId}`"
        :environment-id="environmentId"
        :destination="{ id: selectedId, organizationId }"
        :disabled="!canManage"
      />
      <p v-if="!canManage" class="text-xs text-muted-foreground">
        {{ $t("destinations.readOnly") }}
      </p>
    </template>
    <p v-else class="text-sm text-muted-foreground">{{ $t("destinations.empty") }}</p>
  </section>
</template>
<script setup lang="ts">
import { ref, watch } from "vue";
import {
  listNotificationDestinations,
  listShopNotificationDestinations,
  type NotificationDestination,
} from "@/api/generated";
import NotificationMatrix from "./NotificationMatrix.vue";
const props = defineProps<{
  organizationId: string;
  shopId: number | null;
  environmentId: number;
}>();
const destinations = ref<NotificationDestination[]>([]);
const selectedId = ref("");
const loading = ref(false);
const failed = ref(false);
const canManage = ref(false);
let generation = 0;
async function load() {
  const current = ++generation;
  loading.value = true;
  failed.value = false;
  try {
    const [org, shop] = await Promise.all([
      listNotificationDestinations({ path: { orgId: props.organizationId } }),
      props.shopId
        ? listShopNotificationDestinations({
            path: { orgId: props.organizationId, shopId: props.shopId },
          })
        : Promise.resolve(null),
    ]);
    if (current !== generation) return;
    if (org.error || !org.data || (shop && (shop.error || !shop.data))) {
      failed.value = true;
      return;
    }
    destinations.value = [...org.data.destinations, ...(shop?.data?.destinations ?? [])];
    canManage.value = org.data.canManage && (shop?.data?.canManage ?? true);
    if (!destinations.value.some((d) => d.id === selectedId.value))
      selectedId.value = destinations.value[0]?.id ?? "";
  } catch {
    if (current === generation) failed.value = true;
  } finally {
    if (current === generation) loading.value = false;
  }
}
watch(() => [props.organizationId, props.shopId, props.environmentId], load, { immediate: true });
</script>
