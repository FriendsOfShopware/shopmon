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
            class="group btn"
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
            <form-group
                title="Shop information"
                sub-title=""
            >
                <div class="sm:col-span-6">
                    <label
                        for="Name"
                        class="block text-sm font-medium mb-1"
                    > Name </label>
                    <field
                        id="name"
                        type="text"
                        name="name"
                        autocomplete="name"
                        class="field"
                        :value="shopStore.shop.name"
                        :class="{ 'is-invalid': errors.name }"
                    />
                    <div class="text-red-700">
                        {{ errors.name }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label
                        for="newOrgId"
                        class="block text-sm font-medium mb-1"
                    > Organization </label>
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
                    <div class="text-red-700">
                        {{ errors.newOrgId }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label
                        for="shopUrl"
                        class="block text-sm font-medium mb-1"
                    > URL </label>
                    <field
                        id="shopUrl"
                        type="text"
                        name="shopUrl"
                        autocomplete="url"
                        class="field"
                        :value="shopStore.shop.url"
                        :class="{ 'is-invalid': errors.shopUrl }"
                    />
                    <div class="text-red-700">
                        {{ errors.shopUrl }}
                    </div>
                </div>
            </form-group>

            <form-group
                title="Integration"
                info="<p>The created integration must have access to following
                    <a href='https://github.com/FriendsOfShopware/shopmon/blob/main/app/manifest.xml#L18'>
                        permissions
                    </a>
                </p>"
            >
                <div class="sm:col-span-6">
                    <label
                        for="clientId"
                        class="block text-sm font-medium mb-1"
                    > Client-ID </label>
                    <field
                        id="clientId"
                        type="text"
                        name="clientId"
                        class="field"
                        :class="{ 'is-invalid': errors.clientId }"
                    />
                    <div class="text-red-700">
                        {{ errors.clientId }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label
                        for="clientSecret"
                        class="block text-sm font-medium mb-1"
                    > Client-Secret
                    </label>
                    <field
                        id="clientSecret"
                        type="text"
                        name="clientSecret"
                        class="field"
                        :class="{ 'is-invalid': errors.clientSecret }"
                    />
                    <div class="text-red-700">
                        {{ errors.clientSecret }}
                    </div>
                </div>
            </form-group>

            <div class="flex justify-end pb-12">
                <button
                    :disabled="isSubmitting"
                    type="submit"
                    class="btn btn-primary"
                >
                    <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                        <icon-fa6-solid:floppy-disk
                            v-if="!isSubmitting"
                            class="h-5 w-5"
                            aria-hidden="true"
                        />
                        <icon-line-md:loading-twotone-loop
                            v-else
                            class="w-5 h-5"
                        />
                    </span>
                    Save
                </button>
            </div>
        </vee-form>

        <form-group :title="'Deleting shop ' + shopStore.shop.name">
            <form
                action="#"
                method="POST"
            >
                <p>Once you delete your shop, you will lose all data associated with it. </p>

                <div class="mt-5">
                    <button
                        type="button"
                        class="btn btn-danger group flex items-center"
                        @click="showShopDeletionModal = true"
                    >
                        <icon-fa6-solid:trash class="w-4 h-4 -ml-1 mr-2 opacity-25 group-hover:opacity-50" />
                        Delete shop
                    </button>
                </div>
            </form>
        </form-group>

        <Modal
            :show="showShopDeletionModal"
            close-x-mark
            @close="showShopDeletionModal = false"
        >
            <template #icon>
                <icon-fa6-solid:triangle-exclamation
                    class="h-6 w-6 text-red-600 dark:text-red-400"
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
                    class="btn btn-danger w-full sm:ml-3 sm:w-auto"
                    @click="deleteShop"
                >
                    Delete
                </button>
                <button
                    ref="cancelButtonRef"
                    type="button"
                    class="btn w-full mt-3 sm:w-auto sm:mt-0"
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
