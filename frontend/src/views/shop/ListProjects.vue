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
        <div v-if="projects.length === 0" class="panel">
            <element-empty title="No Projects" button="Add Project" :route="{ name: 'account.projects.new' }">
                Get started by creating your first project and adding shops to it.
            </element-empty>
        </div>

        <div v-else>
            <!-- Projects -->
            <div v-for="project in projects" :key="project.id" class="project-panel panel">
                <div class="panel-header">
                    <div class="project-info">
                        <h3>{{ project.nameCombined }}</h3>

                        <p v-if="project.description" class="project-description">{{ project.description }}</p>

                        <p class="project-meta">
                            <span class="shop-count">{{ projectShops[project.id]?.length || 0 }} shops</span>
                            <span class="separator">•</span>
                            <span class="created-date">Created {{ formatDate(project.createdAt) }}</span>
                        </p>
                    </div>

                    <menu-container as="div" class="menu project-actions">
                        <menu-button>
                            <icon-fa6-solid:ellipsis class="icon menu-button" />
                        </menu-button>

                        <transition
                            enter-active-class="transition ease-out duration-100"
                            enter-from-class="transform opacity-0 scale-95"
                            enter-to-class="transform opacity-100 scale-100"
                            leave-active-class="transition ease-in duration-75"
                            leave-from-class="transform opacity-100 scale-100"
                            leave-to-class="transform opacity-0 scale-95"
                        >
                            <menu-items class="menu-panel">
                                <menu-item>
                                    <button class="menu-item" @click="editProject(project)">
                                        <icon-fa6-solid:pen-to-square class="icon" aria-hidden="true" /> Edit Project
                                    </button>
                                </menu-item>

                                <menu-item>
                                    <router-link class="menu-item" :to="{ name: 'account.shops.new' }">
                                        <icon-fa6-solid:plus class="icon" aria-hidden="true" /> Add Shop
                                    </router-link>
                                </menu-item>

                                <menu-item>
                                    <router-link class="menu-item" :to="{ name: 'account.projects.apikeys', params: { projectId: project.id } }">
                                        <icon-fa6-solid:key class="icon" aria-hidden="true" /> API Keys
                                    </router-link>
                                </menu-item>

                                <menu-item>
                                    <button class="menu-item" :disabled="(projectShops[project.id]?.length || 0) > 0" @click="confirmDeleteProject(project)">
                                        <icon-fa6-solid:trash class="icon" aria-hidden="true" /> Delete Project
                                    </button>
                                </menu-item>
                            </menu-items>
                        </transition>
                    </menu-container>
                </div>
                
                <div v-if="projectShops[project.id]?.length > 0" class="item-grid">
                    <div v-for="shop in projectShops[project.id]" :key="shop.id" class="item">
                        <router-link
                            :to="{
                                    name: 'account.shops.detail',
                                    params: {
                                        slug: organizationSlug,
                                        shopId: shop.id
                                    }
                                }"
                            class="item-link item-wrapper"
                        >
                            <div class="item-logo">
                                <img v-if="shop.favicon" :src="shop.favicon" alt="Shop favicon" class="item-logo-img">
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
                </div>
                
                <element-empty v-else :route="{ name: 'account.shops.new', query: { projectId: project.id } }" title="" button="Add Shop">
                    No shops in this project yet.
                </element-empty>
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
                        autocomplete="off"
                        :class="{ 'has-error': errors.name }"
                    />

                    <div class="field-error-message">
                        {{ errors.name }}
                    </div>
                </div>

                <div class="mt-1">
                    <label for="description">Description</label>

                    <field
                        id="description"
                        v-slot="{ field }"
                        name="description"
                        type="textarea"
                    >
                        <textarea
                            v-bind="field"
                            id="description"
                            autocomplete="off"
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

    <!-- Delete Project Modal -->
    <delete-confirmation-modal
        :show="showDeleteModal"
        title="Delete Project"
        :entity-name="deletingProject?.name || 'this project'"
        @close="showDeleteModal = false"
        @confirm="deleteProject"
    />
</template>

<script setup lang="ts">
import ElementEmpty from '@/components/layout/ElementEmpty.vue';
import Modal from '@/components/layout/Modal.vue';
import DeleteConfirmationModal from '@/components/modal/DeleteConfirmationModal.vue';
import { useAlert } from '@/composables/useAlert';
import { formatDate } from '@/helpers/formatter';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { Field, Form as VeeForm } from 'vee-validate';
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import * as Yup from 'yup';
import { Menu as MenuContainer, MenuButton, MenuItem, MenuItems, } from '@headlessui/vue';

const route = useRoute();
const alert = useAlert();

const loading = ref(true);
const shops = ref<RouterOutput['account']['currentUserShops']>([]);
const projects = ref<RouterOutput['account']['currentUserProjects']>([]);
const editModalVisible = ref(false);
const editingProject = ref({ id: 0, name: '', description: '' });
const showDeleteModal = ref(false);
const deletingProject = ref<(typeof projects.value)[0] | null>(null);

const editSchema = Yup.object().shape({
    name: Yup.string().required('Project name is required'),
    description: Yup.string().optional(),
});

// Get organization slug from the first shop or first project
const organizationSlug = computed(() => {
    if (shops.value.length > 0) {
        return shops.value[0].organizationSlug;
    }
    if (projects.value.length > 0) {
        return projects.value[0].organizationSlug;
    }
    return (route.params.slug as string) || '';
});

// Helper function to get organization ID for a specific project
function getOrganizationIdForProject(projectId: number): string {
    const project = projects.value.find(p => p.id === projectId);
    return project?.organizationId ?? '';
}

// Group shops by project
const projectShops = computed(() => {
    const grouped: Record<number, typeof shops.value> = {};

    for (const shop of shops.value) {
        if (shop.projectId) {
            grouped[shop.projectId] ??= [];
            grouped[shop.projectId].push(shop);
        }
    }

    return grouped;
});

// Load projects and shops in parallel
Promise.all([
    trpcClient.account.currentUserProjects.query(),
    trpcClient.account.currentUserShops.query()
]).then(([projectsData, shopsData]) => {
    projects.value = projectsData;
    shops.value = shopsData;
    loading.value = false;
}).catch(() => {
    loading.value = false;
});

// Edit project
function editProject(project: (typeof projects.value)[0]) {
    editingProject.value = {
        id: project.id,
        name: project.name,
        description: project.description ?? '',
    };
    editModalVisible.value = true;
}

// Update project
async function updateProject(values: Record<string, unknown>) {
    try {
        const typedValues = values as Yup.InferType<typeof editSchema>;
        const orgId = getOrganizationIdForProject(editingProject.value.id);
        
        await trpcClient.organization.project.update.mutate({
            orgId: orgId,
            projectId: editingProject.value.id,
            name: typedValues.name,
            description: typedValues.description ?? '',
        });

        projects.value = await trpcClient.account.currentUserProjects.query();

        editModalVisible.value = false;
        alert.success('Project updated successfully');
    } catch (error) {
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        alert.error(`Failed to update project: ${errorMessage}`);
    }
}

// Confirm delete project
function confirmDeleteProject(project: (typeof projects.value)[0]) {
    deletingProject.value = project;
    showDeleteModal.value = true;
}

// Delete project
async function deleteProject() {
    if (!deletingProject.value) return;

    try {
        const orgId = getOrganizationIdForProject(deletingProject.value.id);
        
        await trpcClient.organization.project.delete.mutate({
            orgId: orgId,
            projectId: deletingProject.value.id,
        });

        projects.value = await trpcClient.account.currentUserProjects.query();
        showDeleteModal.value = false;
        deletingProject.value = null;

        alert.success('Project deleted successfully');
    } catch (error) {
        const errorMessage =
            error instanceof Error ? error.message : 'Failed to delete project';
        alert.error(errorMessage);
    }
}
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

.project-actions {
    .menu-button {
        color: var(--text-color)
    }
}

.project-panel {
    .item-info {
        padding-right: 1.75rem;
    }
}
</style>
