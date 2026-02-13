<template>
  <div class="deployments-container">
    <div class="panel panel-table">
      <div class="panel-header">
        <h3>Product Tokens</h3>
        <button class="btn btn-primary btn-sm" @click="showCreateTokenDialog = true">
          <icon-fa6-solid:plus class="icon" />
          Create Token
        </button>
      </div>

      <data-table
        v-if="tokens && tokens.length > 0"
        :columns="[
          { key: 'name', name: 'Name', class: 'token-name-col', sortable: true },
          { key: 'scope', name: 'Scope', class: 'token-scope-col', sortable: true },
          { key: 'createdAt', name: 'Created', class: 'token-created-col', sortable: true },
          { key: 'lastUsedAt', name: 'Last Used', class: 'token-used-col', sortable: true },
        ]"
        :data="tokens"
      >
        <template #cell-name="{ row }">
          {{ row.name }}
        </template>

        <template #cell-scope="{ row }">
          <span class="scope-badge">{{ row.scope }}</span>
        </template>

        <template #cell-createdAt="{ row }">
          {{ formatDateTime(row.createdAt) }}
        </template>

        <template #cell-lastUsedAt="{ row }">
          {{ row.lastUsedAt ? formatDateTime(row.lastUsedAt) : "Never" }}
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button class="btn btn-danger btn-sm" @click="confirmDeleteToken(row)">
              <icon-fa6-solid:trash class="icon" />
            </button>
          </div>
        </template>
      </data-table>
      <div v-else class="empty-state">
        <icon-fa6-solid:key class="empty-icon" />
        <p>No product tokens created yet.</p>
      </div>
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
              v-if="row.reference"
              :href="row.reference"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-sm btn-secondary"
              title="Open reference link"
            >
              <icon-fa6-solid:arrow-up-right-from-square class="icon" />
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
              <icon-fa6-solid:eye class="icon" />
              View
            </router-link>
            <button class="btn btn-danger btn-sm" @click="confirmDeleteDeployment(row)">
              <icon-fa6-solid:trash class="icon" />
            </button>
          </div>
        </template>
      </data-table>
      <div v-else class="empty-state">
        <icon-fa6-solid:rocket class="empty-icon" />
        <p>No deployments recorded yet.</p>
      </div>
    </div>

    <!-- Create Token Modal -->
    <modal :show="showCreateTokenDialog" close-x-mark @close="showCreateTokenDialog = false">
      <template #icon>
        <icon-fa6-solid:key class="icon icon-info" />
      </template>

      <template #title> Create Product Token </template>

      <template #content>
        <div v-if="!newToken" class="form-group">
          <label for="tokenName">Token Name</label>
          <input
            id="tokenName"
            v-model="tokenName"
            type="text"
            class="field"
            placeholder="e.g., Production Server"
          />
        </div>
        <div v-else class="token-created">
          <alert type="success">
            Token created successfully! Copy it now - you won't be able to see it again.
          </alert>
          <div class="token-display">
            <code>{{ newToken }}</code>
            <button class="btn btn-sm" @click="copyToken">
              <icon-fa6-solid:copy class="icon" />
              Copy
            </button>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          v-if="!newToken"
          type="button"
          class="btn btn-primary"
          :disabled="!tokenName || isCreatingToken"
          @click="createToken"
        >
          Create Token
        </button>
        <button v-else type="button" class="btn btn-primary" @click="closeCreateTokenDialog">
          Close
        </button>
      </template>
    </modal>

    <!-- Delete Token Confirmation Modal -->
    <modal :show="showDeleteTokenDialog" close-x-mark @close="showDeleteTokenDialog = false">
      <template #icon>
        <icon-fa6-solid:triangle-exclamation class="icon icon-error" />
      </template>

      <template #title> Delete Product Token </template>

      <template #content>
        Are you sure you want to delete the token "{{ tokenToDelete?.name }}"? This action cannot be
        undone.
      </template>

      <template #footer>
        <button
          type="button"
          class="btn btn-danger"
          :disabled="isDeletingToken"
          @click="deleteToken"
        >
          Delete
        </button>
        <button type="button" class="btn" @click="showDeleteTokenDialog = false">Cancel</button>
      </template>
    </modal>

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
import { ref, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { useShopDetail } from "@/composables/useShopDetail";
import { formatDateTime } from "@/helpers/formatter";
import { trpcClient } from "@/helpers/trpc";
import Alert from "@/components/layout/Alert.vue";
import Modal from "@/components/layout/Modal.vue";

const route = useRoute();
const { shop } = useShopDetail();

const deployments = ref<any[]>([]);
const tokens = ref<any[]>([]);
const showCreateTokenDialog = ref(false);
const showDeleteTokenDialog = ref(false);
const showDeleteDeploymentDialog = ref(false);
const tokenName = ref("");
const newToken = ref("");
const isCreatingToken = ref(false);
const isDeletingToken = ref(false);
const isDeletingDeployment = ref(false);
const tokenToDelete = ref<any>(null);
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

const loadTokens = async () => {
  if (!shop.value) return;

  try {
    tokens.value = await trpcClient.organization.deployment.listTokens.query({
      shopId: shop.value.id,
    });
  } catch (error) {
    console.error("Failed to load tokens:", error);
  }
};

const createToken = async () => {
  if (!shop.value || !tokenName.value) return;

  isCreatingToken.value = true;
  try {
    const result = await trpcClient.organization.deployment.createToken.mutate({
      shopId: shop.value.id,
      name: tokenName.value,
    });

    newToken.value = result.token;
    await loadTokens();
  } finally {
    isCreatingToken.value = false;
  }
};

const confirmDeleteToken = (token: any) => {
  tokenToDelete.value = token;
  showDeleteTokenDialog.value = true;
};

const deleteToken = async () => {
  if (!shop.value || !tokenToDelete.value) return;

  isDeletingToken.value = true;
  try {
    await trpcClient.organization.deployment.deleteToken.mutate({
      shopId: shop.value.id,
      tokenId: tokenToDelete.value.id,
    });

    await loadTokens();
    showDeleteTokenDialog.value = false;
    tokenToDelete.value = null;
  } finally {
    isDeletingToken.value = false;
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

const closeCreateTokenDialog = () => {
  showCreateTokenDialog.value = false;
  tokenName.value = "";
  newToken.value = "";
};

const copyToken = async () => {
  await navigator.clipboard.writeText(newToken.value);
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
      loadTokens();
    }
  },
  { immediate: true },
);

onMounted(() => {
  if (shop.value) {
    loadDeployments();
    loadTokens();
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

.token-name-col {
  width: 250px;
}

.token-created-col {
  width: 180px;
}

.token-scope-col {
  width: 130px;
}

.token-used-col {
  width: 180px;
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

.token-created {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.token-display {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.token-display code {
  flex: 1;
  padding: 0.75rem;
  background: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-family: monospace;
  word-break: break-all;
}

.scope-badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 3px;
  font-size: 0.8rem;
  font-weight: 500;
  text-transform: capitalize;
  background: var(--surface-color);
  border: 1px solid var(--border-color);
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}
</style>
