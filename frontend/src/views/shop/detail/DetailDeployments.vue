<template>
  <div class="deployments-container">
    <div v-if="shop" class="panel setup-card">
      <h3 class="setup-title">How to Track Deployments</h3>
      <p class="setup-description">
        Use the
        <a
          href="https://github.com/FriendsOfShopware/shopmon-cli"
          target="_blank"
          rel="noopener noreferrer"
          >shopmon-cli</a
        >
        to report deployments. Set the following environment variables and wrap your deployment
        command:
      </p>

      <div class="setup-env">
        <div class="env-row">
          <span class="env-key">SHOPMON_SHOP_ID</span>
          <code class="env-value">{{ shop.id }}</code>
        </div>
        <div class="env-row">
          <span class="env-key">SHOPMON_API_KEY</span>
          <span class="env-value env-placeholder">
            <router-link
              :to="{ name: 'account.projects.apikeys', params: { projectId: shop.projectId } }"
            >
              Manage API Keys
            </router-link>
          </span>
        </div>
      </div>

      <div class="setup-command">
        <code>{{ cliCommand }}</code>
      </div>

      <p class="setup-hint">
        The CLI automatically uses the local <code>git HEAD</code> commit SHA to link deployments to
        your git repository. If that is not available, you can set
        <code>SHOPMON_DEPLOYMENT_VERSION_REFERENCE</code> to specify it manually. The target git
        repository URL can be configured in the
        <router-link :to="{ name: 'account.project.list' }">project settings</router-link>.
      </p>
    </div>

    <div class="panel panel-table">
      <div class="panel-header">
        <h3>Deployment History</h3>
      </div>

      <data-table
        v-if="deployments && deployments.length > 0"
        :columns="[
          { key: 'name', name: 'Name', class: 'deployment-name', sortable: true },
          { key: 'createdAt', name: 'Date', class: 'deployment-date', sortable: true },
          { key: 'returnCode', name: 'Status', class: 'deployment-status', sortable: true },
          { key: 'executionTime', name: 'Duration', class: 'deployment-duration', sortable: true },
        ]"
        :data="deployments"
      >
        <template #cell-name="{ row }">
          <code class="deployment-name-display">{{ row.name }}</code>
        </template>

        <template #cell-createdAt="{ row }">
          {{ formatDateTime(row.createdAt) }}
        </template>

        <template #cell-returnCode="{ row }">
          <span :class="['status-badge', row.returnCode === 0 ? 'status-success' : 'status-error']">
            <icon-fa6-solid:check v-if="row.returnCode === 0" class="icon" />
            <icon-fa6-solid:xmark v-else class="icon" />
            {{ row.returnCode === 0 ? "Success" : `Failed (${row.returnCode})` }}
          </span>
        </template>

        <template #cell-executionTime="{ row }">
          {{ formatDuration(row.executionTime) }}
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <a
              v-if="row.reference && row.gitUrl"
              :href="`${row.gitUrl}/commit/${row.reference}`"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary"
              title="Open commit"
            >
              <icon-fa6-solid:arrow-up-right-from-square />
            </a>
            <router-link
              :to="{
                name: 'account.shops.detail.deployment',
                params: {
                  slug: $route.params.slug,
                  shopId: shop.id,
                  deploymentId: row.id,
                },
              }"
              class="btn btn-sm"
            >
              <icon-fa6-solid:eye />
              View
            </router-link>
            <button class="btn btn-danger" @click="confirmDeleteDeployment(row)">
              <icon-fa6-solid:trash />
            </button>
          </div>
        </template>
      </data-table>
      <div v-else class="empty-state">
        <icon-fa6-solid:rocket class="empty-icon" />
        <p>No deployments recorded yet.</p>
      </div>
    </div>

    <!-- Delete Deployment Confirmation Modal -->
    <modal
      :show="showDeleteDeploymentDialog"
      close-x-mark
      @close="showDeleteDeploymentDialog = false"
    >
      <template #icon>
        <icon-fa6-solid:triangle-exclamation class="icon icon-error" />
      </template>

      <template #title> Delete Deployment </template>

      <template #content>
        Are you sure you want to delete this deployment? This action cannot be undone.
      </template>

      <template #footer>
        <button
          type="button"
          class="btn btn-danger"
          :disabled="isDeletingDeployment"
          @click="deleteDeployment"
        >
          Delete
        </button>
        <button type="button" class="btn" @click="showDeleteDeploymentDialog = false">
          Cancel
        </button>
      </template>
    </modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useShopDetail } from "@/composables/useShopDetail";
import { formatDateTime } from "@/helpers/formatter";
import { trpcClient } from "@/helpers/trpc";
import Modal from "@/components/layout/Modal.vue";

const route = useRoute();
const { shop } = useShopDetail();

const cliCommand = computed(() => {
  const parts = [];
  if (window.location.host !== "shopmon.fos.gg") {
    parts.push(`SHOPMON_BASE_URL=${window.location.protocol}//${window.location.host}`);
  }
  parts.push(`SHOPMON_SHOP_ID=${shop.value?.id ?? ""}`);
  parts.push("SHOPMON_API_KEY=your-api-key");
  parts.push("./shopmon-cli deploy -- vendor/bin/shopware-deployment-helper run");
  return parts.join(" ");
});

const deployments = ref<any[]>([]);
const showDeleteDeploymentDialog = ref(false);
const isDeletingDeployment = ref(false);
const deploymentToDelete = ref<any>(null);

const loadDeployments = async () => {
  if (!shop.value) return;

  try {
    deployments.value = await trpcClient.organization.deployment.list.query({
      shopId: shop.value.id,
      limit: 50,
      offset: 0,
    });
  } catch (error) {
    console.error("Failed to load deployments:", error);
  }
};

const confirmDeleteDeployment = (deployment: any) => {
  deploymentToDelete.value = deployment;
  showDeleteDeploymentDialog.value = true;
};

const deleteDeployment = async () => {
  if (!shop.value || !deploymentToDelete.value) return;

  isDeletingDeployment.value = true;
  try {
    await trpcClient.organization.deployment.delete.mutate({
      shopId: shop.value.id,
      deploymentId: deploymentToDelete.value.id,
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
  shop,
  (newShop) => {
    if (newShop) {
      loadDeployments();
    }
  },
  { immediate: true },
);

onMounted(() => {
  if (shop.value) {
    loadDeployments();
  }
});
</script>

<style scoped>
.deployments-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
}

.panel-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.deployment-name {
  width: 200px;
}

.deployment-name-display {
  font-family: monospace;
  font-size: 0.9rem;
  color: var(--text-primary);
  background: var(--surface-color);
  padding: 0.25rem 0.5rem;
  border-radius: 3px;
}

.deployment-date {
  width: 180px;
}

.deployment-status {
  width: 150px;
}

.deployment-duration {
  width: 100px;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.deployment-command {
  background: var(--surface-color);
  padding: 0.25rem 0.5rem;
  border-radius: 3px;
  font-family: monospace;
  font-size: 0.875rem;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  border-radius: 3px;
  font-size: 0.875rem;
  font-weight: 500;
}

.status-success {
  background: var(--success-bg);
  color: var(--success-color);
}

.status-error {
  background: var(--error-bg);
  color: var(--error-color);
}

.empty-state {
  padding: 3rem;
  text-align: center;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 3rem;
  opacity: 0.3;
  margin-bottom: 1rem;
}

.setup-card {
  padding: 1.25rem;
}

.setup-title {
  margin: 0 0 0.5rem 0;
  font-size: 1.1rem;
  font-weight: 600;
}

.setup-description {
  margin: 0 0 1rem 0;
  color: var(--text-muted);
  font-size: 0.9rem;
  line-height: 1.5;
}

.setup-description a {
  color: var(--primary-color);
  text-decoration: none;
}

.setup-description a:hover {
  text-decoration: underline;
}

.setup-env {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.env-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.env-key {
  font-family: monospace;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-primary);
  min-width: 180px;
}

.env-value {
  font-family: monospace;
  font-size: 0.85rem;
  color: var(--text-muted);
  background: var(--surface-color);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
}

.env-placeholder a {
  color: var(--primary-color);
  text-decoration: none;
}

.env-placeholder a:hover {
  text-decoration: underline;
}

.setup-command {
  background: #0d1117;
  padding: 0.75rem 1rem;
  border-radius: 6px;
}

.setup-hint {
  margin: 0.75rem 0 0 0;
  font-size: 0.8125rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.setup-hint a {
  color: var(--primary-color);
  text-decoration: none;
}

.setup-hint a:hover {
  text-decoration: underline;
}

.setup-hint code {
  font-size: 0.8125rem;
  background: var(--surface-color);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
}

.setup-command code {
  font-family: "SF Mono", "Monaco", "Inconsolata", "Fira Code", monospace;
  font-size: 0.8125rem;
  color: #c9d1d9;
}
</style>
