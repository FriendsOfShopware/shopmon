<template>
  <div v-if="environment" class="space-y-6">
    <!-- Summary cards -->
    <div v-if="deployments.length > 0" class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard :icon="IconRocket" :value="deployments.length" label="Total" />
      <StatCard
        :icon="IconCircleCheck"
        :value="counts.success"
        label="Successful"
        color="success"
      />
      <StatCard
        :icon="IconCircleXmark"
        :value="counts.failed"
        label="Failed"
        :color="counts.failed > 0 ? 'destructive' : 'muted'"
      />
      <StatCard
        v-if="latestDeployment"
        :icon="IconClock"
        :value="formatDateTime(latestDeployment.createdAt)"
        label="Last deploy"
      />
    </div>

    <!-- Deployment list -->
    <div v-if="deployments.length > 0" class="space-y-2">
      <div
        v-for="dep in deployments"
        :key="dep.id"
        :class="[
          'flex items-center gap-4 rounded-xl border px-4 py-3 transition-colors',
          dep.returnCode !== 0 ? 'border-destructive/20 bg-destructive/5' : '',
        ]"
      >
        <!-- Status -->
        <div class="shrink-0">
          <div
            v-if="dep.returnCode === 0"
            class="flex size-8 items-center justify-center rounded-lg bg-success/10"
          >
            <icon-fa6-solid:check class="size-3.5 text-success" />
          </div>
          <div v-else class="flex size-8 items-center justify-center rounded-lg bg-destructive/10">
            <icon-fa6-solid:xmark class="size-3.5 text-destructive" />
          </div>
        </div>

        <!-- Name + date -->
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <RouterLink
              :to="{
                name: 'account.environments.detail.deployment',
                params: {
                  environmentId: environment.id,
                  deploymentId: dep.id,
                },
              }"
              class="truncate font-mono text-sm font-medium hover:text-primary transition-colors"
            >
              {{ dep.name }}
            </RouterLink>
            <Badge
              v-if="dep.returnCode === 0"
              variant="secondary"
              class="inline-flex gap-1 text-[10px]"
            >
              <icon-fa6-solid:check class="size-2" />
              success
            </Badge>
            <Badge v-else variant="destructive" class="inline-flex gap-1 text-[10px]">
              <icon-fa6-solid:xmark class="size-2" />
              exit {{ dep.returnCode }}
            </Badge>
          </div>
          <div class="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
            <span class="tabular-nums">{{ formatDateTime(dep.createdAt) }}</span>
            <span class="tabular-nums">{{ formatDuration(dep.executionTime) }}</span>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex shrink-0 items-center gap-1">
          <Button
            v-if="dep.reference && dep.gitUrl"
            as-child
            variant="ghost"
            size="icon"
            class="size-7"
            :title="$t('deployments.openCommit')"
          >
            <a
              :href="`${dep.gitUrl}/commit/${dep.reference}`"
              target="_blank"
              rel="noopener noreferrer"
            >
              <icon-fa6-solid:code-branch class="size-3.5" />
            </a>
          </Button>
          <Button as-child variant="ghost" size="icon" class="size-7" title="View details">
            <RouterLink
              :to="{
                name: 'account.environments.detail.deployment',
                params: {
                  environmentId: environment.id,
                  deploymentId: dep.id,
                },
              }"
            >
              <icon-fa6-solid:eye class="size-3.5" />
            </RouterLink>
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class="size-7 text-muted-foreground hover:text-destructive"
            title="Delete"
            @click="confirmDeleteDeployment(dep)"
          >
            <icon-fa6-solid:trash class="size-3" />
          </Button>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <EmptyState
      v-else
      :icon="IconRocket"
      :title="$t('deployments.noDeployments')"
      description="Set up the CLI to start tracking deployments."
      size="sm"
    />

    <!-- Setup instructions (collapsible) -->
    <Card>
      <button
        class="flex w-full items-center justify-between px-6 py-4 text-left"
        @click="showSetup = !showSetup"
      >
        <div class="flex items-center gap-2">
          <icon-fa6-solid:terminal class="size-4 text-muted-foreground" />
          <span class="font-semibold">{{ $t("deployments.setupTitle") }}</span>
        </div>
        <icon-fa6-solid:chevron-down
          :class="[
            'size-3 text-muted-foreground transition-transform',
            showSetup ? 'rotate-180' : '',
          ]"
        />
      </button>

      <div v-if="showSetup" class="border-t px-6 pb-5 pt-4 space-y-4">
        <p class="text-sm text-muted-foreground leading-relaxed">
          Use the
          <a
            href="https://github.com/FriendsOfShopware/shopmon-cli"
            target="_blank"
            rel="noopener noreferrer"
            class="text-primary hover:underline"
            >shopmon-cli</a
          >
          to report deployments.
        </p>

        <div class="flex flex-col gap-2">
          <div class="flex items-center gap-3">
            <code class="font-mono text-xs font-semibold min-w-[180px]">SHOPMON_SHOP_ID</code>
            <Badge variant="secondary" class="font-mono text-xs">{{ environment.id }}</Badge>
          </div>
          <div class="flex items-center gap-3">
            <code class="font-mono text-xs font-semibold min-w-[180px]">SHOPMON_API_KEY</code>
            <Button as-child variant="outline" size="sm" class="h-6 text-xs">
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
          </div>
        </div>

        <div class="overflow-x-auto rounded-md bg-[#0d1117] px-4 py-3">
          <code class="font-mono text-[0.8125rem] whitespace-pre text-[#c9d1d9]">{{
            cliCommand
          }}</code>
        </div>

        <p class="text-xs text-muted-foreground leading-relaxed">
          The CLI uses the local
          <code class="rounded bg-muted px-1.5 py-0.5 text-xs">git HEAD</code> commit SHA. Set
          <code class="rounded bg-muted px-1.5 py-0.5 text-xs"
            >SHOPMON_DEPLOYMENT_VERSION_REFERENCE</code
          >
          to override. Configure the git URL in
          <RouterLink :to="{ name: 'account.shop.list' }" class="text-primary hover:underline">{{
            $t("deployments.shopSettings")
          }}</RouterLink
          >.
        </p>
      </div>
    </Card>

    <!-- Delete confirmation -->
    <Dialog
      :open="showDeleteDeploymentDialog"
      @update:open="(v: boolean) => !v && (showDeleteDeploymentDialog = false)"
    >
      <DialogContent class="max-w-md">
        <DialogHeader>
          <div class="flex items-start gap-3">
            <icon-fa6-solid:triangle-exclamation class="mt-0.5 size-5 shrink-0 text-destructive" />
            <DialogTitle>{{ $t("deployments.deleteDeployment") }}</DialogTitle>
          </div>
        </DialogHeader>
        <p>{{ $t("deployments.deleteDeploymentConfirm") }}</p>
        <DialogFooter>
          <Button variant="outline" @click="showDeleteDeploymentDialog = false">{{
            $t("common.cancel")
          }}</Button>
          <Button variant="destructive" :disabled="isDeletingDeployment" @click="deleteDeployment">
            {{ $t("common.delete") }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { formatDateTime } from "@/helpers/formatter";
import { api } from "@/helpers/api";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import StatCard from "@/components/StatCard.vue";
import EmptyState from "@/components/EmptyState.vue";

import IconRocket from "~icons/fa6-solid/rocket";
import IconCircleCheck from "~icons/fa6-solid/circle-check";
import IconCircleXmark from "~icons/fa6-solid/circle-xmark";
import IconClock from "~icons/fa6-solid/clock";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

const route = useRoute();
const { environment } = useEnvironmentDetail();

const showSetup = ref(false);

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

const counts = computed(() => ({
  success: deployments.value.filter((d) => d.returnCode === 0).length,
  failed: deployments.value.filter((d) => d.returnCode !== 0).length,
}));

const latestDeployment = computed(() => deployments.value[0] ?? null);

const loadDeployments = async () => {
  if (!environment.value) return;
  try {
    const { data } = await api.GET("/environments/{environmentId}/deployments", {
      params: { path: { environmentId: environment.value.id }, query: { limit: 50, offset: 0 } },
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
  if (num < 1) return `${(num * 1000).toFixed(0)}ms`;
  if (num < 60) return `${num.toFixed(1)}s`;
  return `${Math.floor(num / 60)}m ${Math.round(num % 60)}s`;
};

watch(
  environment,
  (v) => {
    if (v) loadDeployments();
  },
  { immediate: true },
);
</script>
