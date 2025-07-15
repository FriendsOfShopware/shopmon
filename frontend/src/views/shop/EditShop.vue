<template>
    <header-container
        v-if="shop"
        :title="'Edit ' + shop.name"
    >
        <router-link
            :to="{
                name: 'account.shops.detail',
                params: {
                    slug: route.params.slug as string,
                    shopId: shop.id
                }
            }"
            type="button"
            class="btn"
        >
            Cancel
        </router-link>
    </header-container>

    <main-container v-if="shop">
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="shop"
            class="panel"
            @submit="onSubmit"
        >
            <form-group title="Shop information">
                <div>
                    <label for="Name">Name</label>

                    <field
                        id="name"
                        type="text"
                        name="name"
                        autocomplete="name"
                        class="field"
                        :value="shop.name"
                        :class="{ 'has-error': errors.name }"
                    />

                    <div class="field-error-message">
                        {{ errors.name }}
                    </div>
                </div>

                <div>
                    <label for="projectId">Project</label>

                    <field
                        id="projectId"
                        name="projectId"
                    >
                        <select 
                            v-model="selectedProjectId" 
                            class="field"
                        >
                            <option
                                v-for="project in projects"
                                :key="project.id"
                                :value="project.id"
                            >
                                {{ project.name }}
                            </option>
                        </select>
                    </field>

                    <div class="field-error-message">
                        {{ errors.projectId }}
                    </div>
                </div>

                <div>
                    <label for="shopUrl">URL</label>

                    <field
                        id="shopUrl"
                        type="text"
                        name="shopUrl"
                        autocomplete="url"
                        class="field"
                        :value="shop.url"
                        :class="{ 'has-error': errors.shopUrl }"
                    />

                    <div class="field-error-message">
                        {{ errors.shopUrl }}
                    </div>
                </div>
            </form-group>

            <form-group title="Integration">
                <template #info>
                    The created integration must have access to following
                    <a href="https://github.com/FriendsOfShopware/FroshShopmon?tab=readme-ov-file#permissions">
                        permissions
                    </a>
                </template>

                <div>
                    <label for="clientId">Client-ID</label>

                    <field
                        id="clientId"
                        type="text"
                        name="clientId"
                        class="field"
                        :class="{ 'has-error': errors.clientId }"
                    />

                    <div class="field-error-message">
                        {{ errors.clientId }}
                    </div>
                </div>

                <div>
                    <label for="clientSecret">Client-Secret</label>

                    <field
                        id="clientSecret"
                        type="text"
                        name="clientSecret"
                        class="field"
                        :class="{ 'has-error': errors.clientSecret }"
                    />

                    <div class="field-error-message">
                        {{ errors.clientSecret }}
                    </div>
                </div>
            </form-group>

            <div class="form-submit">
                <button
                    :disabled="isSubmitting"
                    type="submit"
                    class="btn btn-primary"
                >
                    <icon-fa6-solid:floppy-disk
                        v-if="!isSubmitting"
                        class="icon"
                        aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop
                        v-else
                        class="icon"
                    />
                    Save
                </button>
            </div>
        </vee-form>

        <form-group
            v-if="shop"
            title="Sitespeed"
            class="panel"
        >
            <p>
                <a href="https://www.sitespeed.io/" target="_blank">Sitespeed.io</a> allows you to monitor the performance of your shop's frontend.
            </p>
            <p> To activate Sitespeed.io, you need to provide up to five URLs that you want to monitor. These URLs will be used to run performance tests and gather insights about your shop's frontend performance.</p>
            <p>The Sitespeed run is scheduled to run every 24 hours, after the initial run you can view the results of Sitespeed directly in Shopmon. </p>

            <form @submit.prevent="onSitespeedSubmit">
                <div class="mb-1">
                    <label for="sitespeedEnabled">Sitespeed Enabled</label>
                    <input
                        id="sitespeedEnabled"
                        v-model="sitespeedEnabled"
                        type="checkbox"
                        class="field"
                    />
                </div>

                <div v-if="sitespeedEnabled">
                    <label for="sitespeedUrls">Sitespeed URLs</label>

                    <div class="sitespeed-urls-container">
                        <div
                            v-for="(url, index) in sitespeedUrls"
                            :key="index"
                            class="sitespeed-url-row"
                        >
                            <input
                                v-model="sitespeedUrls[index]"
                                type="url"
                                class="field"
                                placeholder="https://example.com"
                            />

                            <button
                                type="button"
                                class="btn btn-icon"
                                @click="removeSitespeedUrl(index)"
                            >
                                <icon-fa6-solid:xmark />
                            </button>
                        </div>

                        <button
                            v-if="sitespeedUrls.length < 5"
                            type="button"
                            class="btn btn-secondary"
                            @click="addSitespeedUrl"
                        >
                            <icon-fa6-solid:plus class="icon" />
                            New URL
                        </button>
                    </div>
                </div>

                <div class="form-submit">
                    <button
                        :disabled="isSitespeedSubmitting || !isSitespeedFormValid"
                        type="submit"
                        class="btn btn-primary"
                    >
                        <icon-fa6-solid:floppy-disk
                            v-if="!isSitespeedSubmitting"
                            class="icon"
                            aria-hidden="true"
                        />

                        <icon-line-md:loading-twotone-loop
                            v-else
                            class="icon"
                        />
                        Save Sitespeed Settings
                    </button>
                </div>
            </form>
        </form-group>

        <form-group :title="'Deleting shop ' + shop.name" class="panel">
                <p>Once you delete your shop, you will lose all data associated with it. </p>

                <button
                    type="button"
                    class="btn btn-danger"
                    @click="showShopDeletionModal = true"
                >
                    <icon-fa6-solid:trash class="icon icon-delete" />
                    Delete shop
                </button>
        </form-group>

        <Modal
            :show="showShopDeletionModal"
            close-x-mark
            @close="showShopDeletionModal = false"
        >
            <template #icon>
                <icon-fa6-solid:triangle-exclamation
                    class="icon icon-error"
                    aria-hidden="true"
                />
            </template>

            <template #title>
                Delete shop
            </template>

            <template #content>
                Are you sure you want to delete your Shop? All of your data will
                be permanently removed
                from our servers forever. This action cannot be undone.
            </template>

            <template #footer>
                <button
                    type="button"
                    class="btn btn-danger"
                    @click="deleteShop"
                >
                    Delete
                </button>

                <button
                    ref="cancelButtonRef"
                    type="button"
                    class="btn"
                    @click="showShopDeletionModal = false"
                >
                    Cancel
                </button>
            </template>
        </Modal>
    </main-container>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';

import { Field, Form as VeeForm } from 'vee-validate';
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import * as Yup from 'yup';

const { error, success } = useAlert();
const router = useRouter();
const route = useRoute();
const shop = ref<RouterOutput['organization']['shop']['get'] | null>(null);
const isLoading = ref(false);
const projects = ref<RouterOutput['account']['currentUserProjects']>([]);
const selectedProjectId = ref<number>(0);

const shopId = Number.parseInt(route.params.shopId as string, 10);

trpcClient.account.currentUserProjects.query().then((data) => {
    projects.value = data;
});

async function loadShop() {
    isLoading.value = true;
    shop.value = await trpcClient.organization.shop.get.query({
        shopId,
    });

    // Load projects for the organization
    if (shop.value?.organizationId) {
        projects.value = await trpcClient.account.currentUserProjects.query();
        selectedProjectId.value = shop.value.projectId;
    }

    // Initialize sitespeed settings
    if (shop.value) {
        sitespeedEnabled.value = shop.value.sitespeedEnabled || false;
        sitespeedUrls.value = shop.value.sitespeedUrls ? [...shop.value.sitespeedUrls] : [];
    }

    isLoading.value = false;
}

loadShop();

const showShopDeletionModal = ref(false);
const sitespeedUrls = ref<string[]>([]);
const sitespeedEnabled = ref(false);
const isSitespeedSubmitting = ref(false);

const isSitespeedFormValid = computed(() => {
    if (!sitespeedEnabled.value) {
        return true; // Always valid when disabled
    }
    // Check if at least one non-empty URL exists
    return sitespeedUrls.value.some(url => url.trim() !== '');
});

const schema = Yup.object().shape({
    name: Yup.string().required('Shop name is required'),
    url: Yup.string().required('Shop URL is required').url(),
    projectId: Yup.number().required('Project is required'),
    clientId: Yup.string().when('url', {
        is: (url: string) => url !== shop.value?.url,
        // biome-ignore lint/suspicious/noThenProperty: Yup schema method
        then: () =>
            Yup.string().required(
                'If you change the URL you need to provide Client-ID',
            ),
    }),
    clientSecret: Yup.string().when('url', {
        is: (url: string) => url !== shop.value?.url,
        // biome-ignore lint/suspicious/noThenProperty: Yup schema method
        then: () =>
            Yup.string().required(
                'If you change the URL you need to provide Client-Secret',
            ),
    }),
});

async function onSubmit(values: Record<string, unknown>) {
    const typedValues = values as Yup.InferType<typeof schema>;
    if (shop.value) {
        try {
            if (typedValues.url) {
                typedValues.url = typedValues.url.replace(/\/+$/, '');
            }
            await trpcClient.organization.shop.update.mutate({
                orgId: shop.value.organizationId,
                shopId: shop.value.id,
                ...typedValues,
                projectId: selectedProjectId.value,
            });

            router.push({
                name: 'account.shops.detail',
                params: {
                    slug: route.params.slug as string,
                    shopId: shop.value.id,
                },
            });
        } catch (e) {
            error(e instanceof Error ? e.message : String(e));
        }
    }
}

async function deleteShop() {
    if (shop.value) {
        try {
            await trpcClient.organization.shop.delete.mutate({
                shopId: shop.value.id,
            });

            router.push({ name: 'account.project.list' });
        } catch (err) {
            error(err instanceof Error ? err.message : String(err));
        }
    }
}

function addSitespeedUrl() {
    sitespeedUrls.value.push('');
}

function removeSitespeedUrl(index: number) {
    sitespeedUrls.value.splice(index, 1);
}

async function onSitespeedSubmit() {
    if (shop.value) {
        try {
            isSitespeedSubmitting.value = true;
            
            // Validate that if enabled, at least one URL is provided
            if (sitespeedEnabled.value && sitespeedUrls.value.length === 0) {
                error('Please provide at least one URL when enabling Sitespeed');
                return;
            }
            
            await trpcClient.organization.shop.updateSitespeedSettings.mutate({
                shopId: shop.value.id,
                enabled: sitespeedEnabled.value,
                urls: sitespeedUrls.value,
            });
            
            // Reload shop data to refresh the UI
            await loadShop();
            success('Sitespeed settings saved successfully');
        } catch (e) {
            error(e instanceof Error ? e.message : String(e));
        } finally {
            isSitespeedSubmitting.value = false;
        }
    }
}
</script>

<style>
.sitespeed-urls-container {
    margin-top: 0.5rem;
}

.sitespeed-url-row {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    align-items: center;
}

.sitespeed-url-row input {
    flex: 1;
}

.btn-icon {
    padding: 0.5rem;
    min-width: auto;
}
</style>
