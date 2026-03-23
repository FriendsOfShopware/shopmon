<template>
  <header-container title="My Projects">
    <div class="header-actions">
      <router-link :to="{ name: 'account.projects.new' }" class="btn btn-primary">
        <icon-fa6-solid:folder-plus class="icon" aria-hidden="true" />
        Add Project
      </router-link>
    </div>
  </header-container>

  <main-container v-if="!loading">
    <Panel v-if="projects.length === 0">
      <element-empty
        title="No Projects"
        button="Add Project"
        :route="{ name: 'account.projects.new' }"
      >
        Get started by creating your first project and adding shops to it.
      </element-empty>
    </Panel>

    <div v-else>
      <!-- Projects -->
      <Panel
        v-for="project in projects"
        :key="project.id"
        class="project-panel"
        :description="project.description || undefined"
      >
        <template #title>{{ project.name }}</template>
        <template #action>
          <router-link
            :to="{ name: 'account.projects.edit', params: { projectId: project.id } }"
            class="btn btn-secondary"
          >
            <icon-fa6-solid:pen-to-square class="icon" aria-hidden="true" />
            Edit Project
          </router-link>
        </template>

        <p class="project-meta">
          <span class="shop-count">{{ projectShops[project.id]?.length || 0 }} shops</span>
        </p>

        <div v-if="projectShops[project.id]?.length > 0" class="item-grid">
          <div v-for="shop in projectShops[project.id]" :key="shop.id" class="item">
            <router-link
              :to="{
                name: 'account.shops.detail',
                params: {
                  organizationId: shop.organizationId,
                  shopId: shop.id,
                },
              }"
              class="item-link item-wrapper"
            >
              <div class="item-logo">
                <img
                  v-if="shop.favicon"
                  :src="shop.favicon"
                  alt="Shop favicon"
                  class="item-logo-img"
                />
                <icon-fa6-solid:store v-else />
              </div>

              <div class="item-info">
                <div class="item-name">
                  {{ shop.name }}
                </div>

                <div class="item-content">{{ shop.shopwareVersion }}</div>
                <status-icon :status="shop.status" class="item-state" />
              </div>
            </router-link>

            <div class="item-actions">
              <a :href="shop.url" target="_blank" class="" data-tooltip="Visit Shop">
                <icon-fa6-solid:arrow-up-right-from-square />
              </a>
            </div>
          </div>

          <router-link
            :to="{ name: 'account.shops.new', query: { projectId: project.id } }"
            class="item item-link item-wrapper add-shop-card"
          >
            <div class="item-logo">
              <icon-fa6-solid:plus />
            </div>
            <div class="item-info">
              <div class="item-name">Add Shop</div>
              <div class="item-content">Add to this project</div>
            </div>
          </router-link>
        </div>

        <element-empty
          v-else
          :route="{ name: 'account.shops.new', query: { projectId: project.id } }"
          title=""
          button="Add Shop"
        >
          No shops in this project yet.
        </element-empty>
      </Panel>
    </div>
  </main-container>
</template>

<script setup lang="ts">
import ElementEmpty from "@/components/layout/ElementEmpty.vue";
import { formatDate } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { computed, ref } from "vue";
import { useAccountShops } from "@/composables/useAccountShops";

const loading = ref(true);
const { shops } = useAccountShops();
const projects = ref<components["schemas"]["AccountProject"][]>([]);

// Group shops by project
const projectShops = computed(() => {
  const grouped: Record<number, typeof shops.value> = {};

  for (const shop of shops.value) {
    const projectId = shop.projectId;
    if (projectId) {
      grouped[projectId] ??= [];
      grouped[projectId].push(shop);
    }
  }

  return grouped;
});

// Load projects
api
  .GET("/account/projects")
  .then((projectsRes) => {
    if (projectsRes.data) projects.value = projectsRes.data;
    loading.value = false;
  })
  .catch(() => {
    loading.value = false;
  });
</script>

<style scoped>
.project-info {
  flex: 1;
}

.project-description {
  margin: 0 0 0.5rem 0;
}

.project-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--text-color-muted);
}

.separator {
  opacity: 0.5;
}

.project-panel {
  .item-info {
    padding-right: 1.75rem;
  }
}

.add-shop-card {
  border: 1px dashed var(--panel-border-color);
  min-height: 70px;

  .item-logo {
    color: var(--text-color-muted);
  }
}
</style>
