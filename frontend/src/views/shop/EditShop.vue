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
            @submit="onSubmit"
            class="panel"
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
                    <a href="https://github.com/FriendsOfShopware/shopmon/blob/main/app/manifest.xml#L18">
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
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import * as Yup from 'yup';

const { error } = useAlert();
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

    isLoading.value = false;
}

loadShop();

const showShopDeletionModal = ref(false);

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
</script>
