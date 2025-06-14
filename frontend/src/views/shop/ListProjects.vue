<template>
    <header-container title="My Projects">
        <div class="header-actions">
            <router-link :to="{ name: 'account.projects.new' }" class="btn btn-primary">
                <icon-fa6-solid:folder-plus class="icon" aria-hidden="true" />
                Add Project
            </router-link>
            <router-link :to="{ name: 'account.shops.new' }" class="btn btn-secondary">
                <icon-fa6-solid:plus class="icon" aria-hidden="true" />
                Add Shop
            </router-link>
        </div>
    </header-container>

    <main-container v-if="!loading">
        <template v-if="projects.length === 0 && shops.length === 0">
            <element-empty title="No Projects" button="Add Project" :route="{ name: 'account.projects.new' }">
                Get started by creating your first project and adding shops to it.
            </element-empty>
        </template>

        <div v-else>
            <!-- Projects -->
            <div v-for="project in projects" :key="project.id" class="project-card">
                <div class="project-header">
                    <div class="project-info">
                        <h3 class="project-name">{{ project.name }}</h3>
                        <p v-if="project.description" class="project-description">{{ project.description }}</p>
                        <p class="project-meta">
                            <span class="shop-count">{{ projectShops[project.id]?.length || 0 }} shops</span>
                            <span class="separator">•</span>
                            <span class="created-date">Created {{ formatDate(project.createdAt) }}</span>
                        </p>
                    </div>
                    <div class="project-actions">
                        <button class="btn btn-sm btn-ghost" data-tooltip="Edit Project" @click="editProject(project)">
                            <icon-fa6-solid:pen-to-square />
                        </button>
                        <button class="btn btn-sm btn-ghost btn-danger" data-tooltip="Delete Project" :disabled="(projectShops[project.id]?.length || 0) > 0" @click="deleteProject(project)">
                            <icon-fa6-solid:trash />
                        </button>
                    </div>
                </div>
                
                <div v-if="projectShops[project.id]?.length > 0" class="project-shops">
                    <div class="shop-grid">
                        <div v-for="shop in projectShops[project.id]" :key="shop.id" class="shop-card">
                            <status-icon :status="shop.status" class="shop-status" />
                            <img v-if="shop.favicon" :src="shop.favicon" alt="Shop favicon" class="shop-favicon">
                            <div v-else class="shop-favicon-placeholder">
                                <icon-fa6-solid:store />
                            </div>
                            <div class="shop-info">
                                <router-link 
                                    :to="{
                                        name: 'account.shops.detail',
                                        params: {
                                            slug: organizationSlug,
                                            shopId: shop.id
                                        }
                                    }" 
                                    class="shop-name"
                                >
                                    {{ shop.name }}
                                </router-link>
                                <div class="shop-version">{{ shop.shopwareVersion }}</div>
                            </div>
                            <div class="shop-actions">
                                <a :href="shop.url" target="_blank" class="btn btn-xs btn-ghost" data-tooltip="Visit Shop">
                                    <icon-fa6-solid:arrow-up-right-from-square />
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
                <div v-else class="no-shops">
                    <p>No shops in this project yet.</p>
                    <router-link :to="{ name: 'account.shops.new', query: { projectId: project.id } }" class="btn btn-sm btn-primary">
                        Add Shop
                    </router-link>
                </div>
            </div>

        </div>
    </main-container>

    <!-- Edit Project Modal -->
    <modal :show="editModalVisible" @close="editModalVisible = false">
        <template #title>
            Edit Project
        </template>
        <template #content>
            <vee-form
                v-slot="{ errors, isSubmitting }"
                :validation-schema="editSchema"
                :initial-values="editingProject"
                @submit="updateProject"
            >
                <div>
                    <label for="name">Project Name</label>

                    <field
                        id="name"
                        type="text"
                        name="name"
                        class="field"
                        :class="{ 'has-error': errors.name }"
                    />

                    <div class="field-error-message">
                        {{ errors.name }}
                    </div>
                </div>

                <div class="form-group">
                    <label for="description">Description</label>

                    <field
                        id="description"
                        v-slot="{ field }"
                        name="description"
                    >
                        <textarea
                            v-bind="field"
                            id="description"
                            class="field"
                            rows="3"
                            placeholder="Optional project description..."
                            :class="{ 'has-error': errors.description }"
                        />
                    </field>

                    <div class="field-error-message">
                        {{ errors.description }}
                    </div>
                </div>

                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" @click="editModalVisible = false">Cancel</button>
                    <button type="submit" class="btn btn-primary" :disabled="isSubmitting">
                        <span v-if="isSubmitting">Saving...</span>
                        <span v-else>Save Changes</span>
                    </button>
                </div>
            </vee-form>
        </template>
    </modal>
</template>

<script setup lang="ts">
import ElementEmpty from '@/components/layout/ElementEmpty.vue';
import Modal from '@/components/layout/Modal.vue';
import { useAlert } from '@/composables/useAlert';
import { formatDate } from '@/helpers/formatter';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { Field, Form as VeeForm } from 'vee-validate';
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import * as Yup from 'yup';

const route = useRoute();
const alert = useAlert();

const loading = ref(true);
const shops = ref<RouterOutput['account']['currentUserShops']>([]);
const projects = ref<RouterOutput['organization']['project']['list']>([]);
const editModalVisible = ref(false);
const editingProject = ref({ id: 0, name: '', description: '' });

const editSchema = Yup.object().shape({
    name: Yup.string().required('Project name is required'),
    description: Yup.string().optional(),
});

// Get organization slug from the first shop or route
const organizationSlug = computed(() => {
    if (shops.value.length > 0) {
        return shops.value[0].organizationSlug;
    }
    return (route.params.slug as string) || '';
});

// Get organization ID from the first shop
const organizationId = computed(() => {
    if (shops.value.length > 0) {
        return shops.value[0].organizationId;
    }
    return '';
});

// Group shops by project
const projectShops = computed(() => {
    const grouped: Record<number, typeof shops.value> = {};

    for (const shop of shops.value) {
        if (shop.projectId) {
            if (!grouped[shop.projectId]) {
                grouped[shop.projectId] = [];
            }
            grouped[shop.projectId].push(shop);
        }
    }

    return grouped;
});

trpcClient.account.currentUserShops.query().then((shopsData) => {
        shops.value = shopsData;

        if (organizationId.value) {
            // Load projects for the organization
            trpcClient.organization.project.list
                .query({ orgId: organizationId.value })
                .then((projectsData) => {
                    projects.value = projectsData;
                    loading.value = false;
                });
        } else {
            loading.value = false;
        }
    })
    .catch(() => {
        loading.value = false;
    });

// Edit project
function editProject(project: (typeof projects.value)[0]) {
    editingProject.value = {
        id: project.id,
        name: project.name,
        description: project.description || '',
    };
    editModalVisible.value = true;
}

// Update project
async function updateProject(values: Record<string, unknown>) {
    try {
        const typedValues = values as Yup.InferType<typeof editSchema>;
        await trpcClient.organization.project.update.mutate({
            orgId: organizationId.value,
            projectId: editingProject.value.id,
            name: typedValues.name,
            description: typedValues.description || '',
        });

        // Reload projects
        const projectsData = await trpcClient.organization.project.list.query({
            orgId: organizationId.value,
        });
        projects.value = projectsData;

        editModalVisible.value = false;
        alert.success('Project updated successfully');
    } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        alert.error(
            `Failed to update project: ${errorMessage}`,
        );
    }
}

// Delete project
async function deleteProject(project: (typeof projects.value)[0]) {
    if (
        !window.confirm(
            `Are you sure you want to delete the project "${project.name}"?`,
        )
    ) {
        return;
    }

    try {
        await trpcClient.organization.project.delete.mutate({
            orgId: organizationId.value,
            projectId: project.id,
        });

        // Reload projects
        const projectsData = await trpcClient.organization.project.list.query({
            orgId: organizationId.value,
        });
        projects.value = projectsData;

        alert.success('Project deleted successfully');
    } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Failed to delete project';
        alert.error(errorMessage);
    }
}
</script>

<style scoped>
.header-actions {
    display: flex;
    gap: 0.5rem;
}

.project-card {
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius-lg);
    padding: 1.5rem;
    margin-bottom: 1.5rem;
}

.project-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
}

.project-info {
    flex: 1;
}

.project-name {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0 0 0.25rem 0;
    color: var(--color-text-primary);
}

.project-description {
    color: var(--color-text-secondary);
    margin: 0 0 0.5rem 0;
}

.project-meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: var(--color-text-muted);
}

.separator {
    opacity: 0.5;
}

.project-actions {
    display: flex;
    gap: 0.25rem;
}

.project-shops {
    margin-top: 1rem;
}

.shop-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
}

.shop-card {
    background: var(--panel-background);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius);
    padding: 1rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    transition: all 0.2s;
}

.shop-card:hover {
    border-color: var(--color-primary);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.shop-status {
    flex-shrink: 0;
}

.shop-favicon {
    width: 32px;
    height: 32px;
    border-radius: var(--border-radius);
    flex-shrink: 0;
}

.shop-favicon-placeholder {
    width: 32px;
    height: 32px;
    border-radius: var(--border-radius);
    background: var(--color-surface);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: var(--color-text-muted);
}

.shop-info {
    flex: 1;
    min-width: 0;
}

.shop-name {
    font-weight: 500;
    color: var(--color-text-primary);
    text-decoration: none;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.shop-name:hover {
    color: var(--color-primary);
}

.shop-version {
    font-size: 0.875rem;
    color: var(--color-text-muted);
}

.shop-actions {
    display: flex;
    gap: 0.25rem;
    flex-shrink: 0;
}

.no-shops {
    text-align: center;
    padding: 2rem;
    color: var(--color-text-muted);
}

.no-shops p {
    margin-bottom: 1rem;
}

.btn-danger:hover {
    color: var(--color-danger);
}

.btn-danger:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.form-group {
    margin-bottom: 1rem;
}

.form-group:last-child {
    margin-bottom: 0;
}

.modal-footer {
    margin-top: 1.5rem;
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
}

.form-label {
    display: block;
    margin-bottom: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-primary);
}

.field {
    width: 100%;
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    color: var(--color-text-primary);
    background-color: var(--color-background);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius);
    transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
}

.field:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px rgba(var(--color-primary-rgb), 0.1);
}

textarea.field {
    resize: vertical;
    min-height: 4rem;
}
</style>
