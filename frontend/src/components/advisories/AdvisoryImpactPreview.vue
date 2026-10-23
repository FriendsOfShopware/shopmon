<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  adminPreviewAdvisoryImpact,
  type AdminAdvisoryImpact,
  type AdvisoryPackageInput,
} from "@/api/generated";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

const props = defineProps<{ packages: AdvisoryPackageInput[] }>();
const { t } = useI18n();
const open = ref(false);
const loading = ref(false);
const error = ref("");
const result = ref<AdminAdvisoryImpact | null>(null);
const checkedPackages = ref<AdvisoryPackageInput[]>([]);
const metrics = computed(() => [
  { label: t("admin.advisoryImpact.shopCount"), value: result.value?.shopCount },
  { label: t("admin.advisoryImpact.environmentCount"), value: result.value?.environmentCount },
]);
let requestId = 0;

async function checkImpact() {
  const request = ++requestId;
  checkedPackages.value = props.packages.map((pkg) => ({
    packageName: pkg.packageName.trim(),
    affectedVersions: pkg.affectedVersions.trim(),
  }));
  open.value = true;
  loading.value = true;
  error.value = "";
  result.value = null;
  try {
    const { data, error: apiError } = await adminPreviewAdvisoryImpact({
      body: { packages: checkedPackages.value },
    });
    if (request !== requestId) return;
    if (apiError || !data) {
      error.value = apiError?.message ?? t("admin.advisoryImpact.failed");
      return;
    }
    result.value = data;
  } catch {
    if (request === requestId) error.value = t("admin.advisoryImpact.failed");
  } finally {
    if (request === requestId) loading.value = false;
  }
}

// Closing, changing constraints, or navigating away invalidates in-flight results.
watch(
  open,
  (value) => {
    if (!value) {
      requestId++;
      loading.value = false;
    }
  },
  { flush: "sync" },
);
watch(
  () => props.packages,
  () => {
    open.value = false;
  },
  { deep: true },
);
onBeforeUnmount(() => {
  requestId++;
});
</script>

<template>
  <Button type="button" variant="outline" :disabled="loading" @click="checkImpact">
    <icon-fa6-solid:store class="mr-1.5 size-3.5" />
    {{ $t("admin.advisoryImpact.button") }}
  </Button>
  <Dialog v-model:open="open">
    <DialogContent class="max-h-[85dvh] overflow-y-auto sm:max-w-3xl">
      <DialogHeader>
        <DialogTitle>{{ $t("admin.advisoryImpact.title") }}</DialogTitle>
        <DialogDescription>{{ $t("admin.advisoryImpact.hint") }}</DialogDescription>
      </DialogHeader>

      <div class="space-y-1 text-xs">
        <p class="font-medium">{{ $t("admin.advisoryImpact.constraints") }}</p>
        <div
          v-for="(pkg, index) in checkedPackages"
          :key="index"
          class="flex flex-wrap gap-x-2 font-mono text-muted-foreground"
        >
          <span>{{ pkg.packageName }}</span
          ><span>{{ pkg.affectedVersions }}</span>
        </div>
      </div>

      <p
        v-if="loading"
        role="status"
        class="flex items-center gap-2 py-6 text-sm text-muted-foreground"
      >
        <icon-line-md:loading-twotone-loop class="size-4" />
        {{ $t("admin.advisoryImpact.loading") }}
      </p>
      <div v-else-if="error" class="space-y-3">
        <Alert variant="destructive"
          ><AlertDescription>{{ error }}</AlertDescription></Alert
        >
        <Button variant="outline" @click="checkImpact">{{
          $t("admin.advisoryImpact.retry")
        }}</Button>
      </div>
      <template v-else-if="result">
        <dl class="grid grid-cols-2 gap-4">
          <div v-for="metric in metrics" :key="metric.label">
            <dt class="text-xs font-medium text-muted-foreground">{{ metric.label }}</dt>
            <dd class="mt-1 text-2xl font-semibold tabular-nums">{{ metric.value }}</dd>
          </div>
        </dl>
        <p class="text-xs text-muted-foreground">{{ $t("admin.advisoryImpact.matchingHint") }}</p>
        <p v-if="result.shops.length === 0" class="py-4 text-sm text-muted-foreground">
          {{ $t("admin.advisoryImpact.empty") }}
        </p>
        <div v-for="shop in result.shops" :key="shop.shopId" class="min-w-0 rounded-lg border">
          <div class="border-b bg-muted/40 px-4 py-3">
            <h3 class="font-medium">{{ shop.shopName }}</h3>
            <RouterLink
              :to="{ name: 'admin.organizations.detail', params: { id: shop.organizationId } }"
              class="text-xs text-muted-foreground hover:underline"
            >
              {{ shop.organizationName }}
            </RouterLink>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="text-muted-foreground">
                <tr>
                  <th class="px-4 py-2 font-medium">{{ $t("dashboard.environment") }}</th>
                  <th class="px-4 py-2 font-medium">{{ $t("admin.packageName") }}</th>
                  <th class="px-4 py-2 font-medium">
                    {{ $t("admin.advisoryImpact.installedVersion") }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="match in shop.matches"
                  :key="`${match.environmentId}:${match.packageName}`"
                  class="border-t"
                >
                  <td class="px-4 py-2">
                    <RouterLink
                      :to="{
                        name: 'admin.environments.detail',
                        params: { id: match.environmentId },
                      }"
                      class="text-primary hover:underline"
                      >{{ match.environmentName }}</RouterLink
                    >
                  </td>
                  <td class="px-4 py-2 font-mono">
                    <div>{{ match.packageName }}</div>
                    <div class="mt-1 text-muted-foreground">{{ match.affectedVersions }}</div>
                  </td>
                  <td class="px-4 py-2 font-mono">{{ match.installedVersion }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </DialogContent>
  </Dialog>
</template>
