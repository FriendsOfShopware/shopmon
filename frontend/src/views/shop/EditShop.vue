<template>
    <header-container
        v-if="shopStore.shop"
        :title="'Edit ' + shopStore.shop.name"
    >
        <router-link
            :to="{
                name: 'account.shops.detail',
                params: {
                    organizationId: shopStore.shop.organizationId,
                    shopId: shopStore.shop.id
                }
            }"
            type="button"
            class="btn"
        >
            Cancel
        </router-link>
    </header-container>

    <main-container v-if="shopStore.shop && authStore.user">
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="shopStore.shop"
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
                        :value="shopStore.shop.name"
                        :class="{ 'has-error': errors.name }"
                    />

                    <div class="field-error-message">
                        {{ errors.name }}
                    </div>
                </div>

                <div>
                    <label for="newOrgId">Organization</label>

                    <field
                        id="newOrgId"
                        as="select"
                        name="newOrgId"
                        class="field"
                        :value="shopStore.shop.organizationId"
                    >
                        <option
                            v-for="organization in authStore.user.organizations"
                            :key="organization.id"
                            :value="organization.id"
                        >
                            {{ organization.name }}
                        </option>
                    </field>

                    <div class="field-error-message">
                        {{ errors.newOrgId }}
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
                        :value="shopStore.shop.url"
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
                    <a href='https://github.com/FriendsOfShopware/shopmon/blob/main/app/manifest.xml#L18'>
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

        <form-group :title="'Deleting shop ' + shopStore.shop.name">
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
import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';

import { Form as VeeForm, Field } from 'vee-validate';
import { useRouter, useRoute } from 'vue-router';
import * as Yup from 'yup';
import { ref } from 'vue';

const authStore = useAuthStore();
const shopStore = useShopStore();
const alertStore = useAlertStore();
const router = useRouter();
const route = useRoute();

const organizationId = parseInt(route.params.organizationId as string, 10);
const shopId = parseInt(route.params.shopId as string, 10);

shopStore.loadShop(organizationId, shopId);

const showShopDeletionModal = ref(false);

const schema = Yup.object().shape({
    name: Yup.string().required('Shop name is required'),
    url: Yup.string().required('Shop URL is required').url(),
    organizationId: Yup.number().required('Organization is required'),
    clientId: Yup.string().when('url', {
        is: (url: string) => url !== shopStore.shop?.url,
        then: () => Yup.string().required('If you change the URL you need to provide Client-ID'),
    }),
    clientSecret: Yup.string().when('url', {
        is: (url: string) => url !== shopStore.shop?.url,
        then: () => Yup.string().required('If you change the URL you need to provide Client-Secret'),
    }),
});

async function onSubmit(values: Yup.InferType<typeof schema>) {
    if (shopStore.shop) {
        try {
            await shopStore.updateShop(shopStore.shop.organizationId, shopStore.shop.id, values);

            router.push({
                name: 'account.shops.detail',
                params: {
                    organizationId: shopStore.shop.organizationId,
                    shopId: shopStore.shop.id,
                },
            });
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}

async function deleteShop() {
    if (shopStore.shop) {
        try {
            await shopStore.delete(shopStore.shop.organizationId, shopStore.shop.id);

            router.push('/account/shops');
        } catch (error: any) {
            alertStore.error(error);
        }
    }
}

</script>
