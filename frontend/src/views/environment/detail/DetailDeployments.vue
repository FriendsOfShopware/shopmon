<template>
  <div class="flex flex-col gap-4">
    <Card v-if="environment">
      <CardContent>
        <h3 class="text-lg font-semibold mb-2">{{ $t("deployments.setupTitle") }}</h3>
        <p class="text-sm text-muted-foreground mb-4 leading-relaxed">
          Use the
          <a
            href="https://github.com/FriendsOfShopware/shopmon-cli"
            target="_blank"
            rel="noopener noreferrer"
            class="text-primary hover:underline"
            >shopmon-cli</a
          >
          to report deployments. Set the following environment variables and wrap your deployment
          command:
        </p>

        <div class="flex flex-col gap-2 mb-4">
          <div class="flex items-center gap-3">
            <span class="font-mono text-sm font-semibold min-w-[180px]">SHOPMON_SHOP_ID</span>
            <code class="font-mono text-sm text-muted-foreground bg-muted px-2 py-0.5 rounded">{{ environment.id }}</code>
          </div>
          <div class="flex items-center gap-3">
            <span class="font-mono text-sm font-semibold min-w-[180px]">SHOPMON_API_KEY</span>
            <span class="font-mono text-sm">
              <Button as-child variant="outline" size="sm">
                <RouterLink
                  :to="{
                    name: 'account.shops.edit',
                    params: { shopId: environment.shopId },
                    hash: '#api-keys',
                  }"
                >
                  {{ $t("deployments.manageApiKeys") }}
                </RouterLink>
              </Button>
            </span>
          </div>
        </div>

        <div class="bg-[#0d1117] px-4 py-3 rounded-md">
          <code class="font-mono text-[0.8125rem] text-[#c9d1d9]">{{ cliCommand }}</code>
        </div>

        <p class="mt-3 text-xs text-muted-foreground leading-relaxed">
          The CLI automatically uses the local <code class="bg-muted px-1.5 py-0.5 rounded text-xs">git HEAD</code> commit SHA to link deployments to
          your git repository. If that is not available, you can set
          <code class="bg-muted px-1.5 py-0.5 rounded text-xs">SHOPMON_DEPLOYMENT_VERSION_REFERENCE</code> to specify it manually. The target git
          repository URL can be configured in the
          <router-link :to="{ name: 'account.shop.list' }" class="text-primary hover:underline">{{
            $t("deployments.shopSettings")
          }}</router-link
          >.
        </p>
      </CardContent>
    </Card>

    <Card class="p-0 overflow-hidden">
      <CardHeader>
        <CardTitle>{{ $t('deployments.history') }}</CardTitle>
      </CardHeader>
      <data-table
        v-if="deployments && deployments.length > 0"
        :columns="[
          { key: 'name', name: $t('common.name'), class: 'w-[200px]', sortable: true },
          { key: 'createdAt', name: $t('common.date'), class: 'w-[180px]', sortable: true },
          {
            key: 'returnCode',
            name: $t('common.status'),
            class: 'w-[150px]',
            sortable: true,
          },
          {
            key: 'executionTime',
            name: $t('deployments.duration'),
            class: 'w-[100px]',
            sortable: true,
          },
        ]"
        :data="deployments"
      >
        <template #cell-name="{ row }">
          <code class="font-mono text-sm bg-muted px-2 py-1 rounded">{{ row.name }}</code>
        </template>

        <template #cell-createdAt="{ row }">
          {{ formatDateTime(row.createdAt) }}
        </template>

        <template #cell-returnCode="{ row }">
          <Badge
            :variant="row.returnCode === 0 ? 'default' : 'destructive'"
            class="gap-1"
            :class="row.returnCode === 0 ? 'bg-green-600' : ''"
          >
            <icon-fa6-solid:check v-if="row.returnCode === 0" class="size-2.5" />
            <icon-fa6-solid:xmark v-else class="size-2.5" />
            {{
              row.returnCode === 0
                ? $t("deployments.success")
                : $t("deployments.failed", { code: row.returnCode })
            }}
          </Badge>
        </template>

        <template #cell-executionTime="{ row }">
          {{ formatDuration(row.executionTime) }}
        </template>

        <template #cell-actions="{ row }">
          <div class="flex gap-2 items-center">
            <Button
              v-if="row.reference && row.gitUrl"
              as-child
              variant="outline"
              size="icon-sm"
            >
              <a
                :href="`${row.gitUrl}/commit/${row.reference}`"
                target="_blank"
                rel="noopener noreferrer"
                :title="$t('deployments.openCommit')"
              >
                <icon-fa6-solid:arrow-up-right-from-square />
              </a>
            </Button>
            <Button as-child variant="outline" size="sm">
              <RouterLink
                :to="{
                  name: 'account.environments.detail.deployment',
                  params: {
                    organizationId: $route.params.organizationId,
                    environmentId: environment?.id ?? 0,
                    deploymentId: row.id,
                  },
                }"
              >
                <icon-fa6-solid:eye />
                {{ $t("common.view") }}
              </RouterLink>
            </Button>
            <Button variant="destructive" size="sm" @click="confirmDeleteDeployment(row)">
              <icon-fa6-solid:trash />
            </Button>
          </div>
        </template>
      </data-table>
      <div v-else class="p-12 text-center text-muted-foreground">
        <icon-fa6-solid:rocket class="size-12 opacity-30 mb-4" />
        <p>{{ $t("deployments.noDeployments") }}</p>
      </div>
    </Card>

    <!-- Delete Deployment Confirmation Modal -->
    <modal
      :show="showDeleteDeploymentDialog"
      close-x-mark
      @close="showDeleteDeploymentDialog = false"
    >
      <template #icon>
        <icon-fa6-solid:triangle-exclamation class="size-6 text-destructive" />
      </template>

      <template #title> {{ $t("deployments.deleteDeployment") }} </template>

      <template #content>
        {{ $t("deployments.deleteDeploymentConfirm") }}
      </template>

      <template #footer>
        <Button
          type="button"
          variant="destructive"
          :disabled="isDeletingDeployment"
          @click="deleteDeployment"
        >
          {{ $t("common.delete") }}
        </Button>
        <Button type="button" variant="outline" @click="showDeleteDeploymentDialog = false">
          {{ $t("common.cancel") }}
        </Button>
      </template>
    </modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { formatDateTime } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import Modal from "@/components/layout/Modal.vue";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

const route = useRoute();
const { environment } = useEnvironmentDetail();

const cliCommand = computed(() => {
  const parts = [];
  if (window.location.host !== "shopmon.fos.gg") {
    parts.push(`SHOPMON_BASE_URL=${window.location.protocol}//${window.location.host}`);
  }
  parts.push(`SHOPMON_SHOP_ID=${environment.value?.id ?? ""}`);
  parts.push("SHOPMON_API_KEY=your-api-key");
  parts.push("./shopmon-cli deploy -- vendor/bin/shopware-deployment-helper run");
  return parts.join(" ");
});

const deployments = ref<any[]>([]);
const showDeleteDeploymentDialog = ref(false);
const isDeletingDeployment = ref(false);
const deploymentToDelete = ref<any>(null);

const loadDeployments = async () => {
  if (!environment.value) return;

  try {
    const { data } = await api.GET("/environments/{environmentId}/deployments", {
      params: {
        path: { environmentId: environment.value.id },
        query: { limit: 50, offset: 0 },
      },
    });
    deployments.value = data ?? [];
  } catch (error) {
    console.error("Failed to load deployments:", error);
  }
};

const confirmDeleteDeployment = (deployment: any) => {
  deploymentToDelete.value = deployment;
  showDeleteDeploymentDialog.value = true;
};

const deleteDeployment = async () => {
  if (!environment.value || !deploymentToDelete.value) return;

  isDeletingDeployment.value = true;
  try {
    await api.DELETE("/environments/{environmentId}/deployments/{deploymentId}", {
      params: {
        path: { environmentId: environment.value.id, deploymentId: deploymentToDelete.value.id },
      },
    });

    await loadDeployments();
    showDeleteDeploymentDialog.value = false;
    deploymentToDelete.value = null;
  } finally {
    isDeletingDeployment.value = false;
  }
};

const formatDuration = (seconds: string) => {
  const num = parseFloat(seconds);
  if (num < 1) {
    return `${(num * 1000).toFixed(0)}ms`;
  }
  return `${num.toFixed(2)}s`;
};

watch(
  environment,
  (newEnvironment) => {
    if (newEnvironment) {
      loadDeployments();
    }
  },
  { immediate: true },
);

onMounted(() => {
  if (environment.value) {
    loadDeployments();
  }
});
</script>
