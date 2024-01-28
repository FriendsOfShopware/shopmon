<script setup lang="ts">
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';

import { trpcClient } from '@/helpers/trpc';
import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useRouter, useRoute } from 'vue-router';

import { useForm } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import * as z from 'zod';

import { ref } from 'vue';

const authStore = useAuthStore();
const alertStore = useAlertStore();
const router = useRouter();
const route = useRoute();

const orgId = parseInt(route.params.organizationId as string, 10);
const shopId = parseInt(route.params.shopId as string, 10);

const shop = await trpcClient.organization.shop.get.query({ orgId, shopId });
const organizations = await trpcClient.account.currentUserOrganizations.query();

const showShopDeletionModal = ref(false);

const { values, handleSubmit, errors, isSubmitting  } = useForm({
    validationSchema: toTypedSchema(z.object({
        name: z.string().min(1, 'Name of organization is required'),
        url: z.string().url('URL of shop is required'),
        newOrgId: z.number().min(1, 'Organization is required'),
        shopUrl: z.string().url('URL of shop is required'),

        // todo: only required if url exists
        clientId: z.string().min(1, 'Client-ID is required'),
        clientSecret: z.string().min(1, 'Client-Secret is required'),
        /*
        clientId: Yup.string().when('url', {
            is: (url: string) => url !== shopStore.shop?.url,
            then: () => Yup.string().required('If you change the URL you need to provide Client-ID'),
        }),
        clientSecret: Yup.string().when('url', {
            is: (url: string) => url !== shopStore.shop?.url,
            then: () => Yup.string().required('If you change the URL you need to provide Client-Secret'),
        }),
        */
    })),
    initialValues: shop,
});


const onSubmit = handleSubmit(async (values) => {
    try {
        values.shopUrl = values.shopUrl.replace(/\/+$/, '');
            
        await trpcClient.organization.shop.update.mutate({ orgId, shopId, ...values });

        router.push({
            name: 'account.shops.detail',
            params: {
                organizationId: orgId,
                shopId: shopId,
            },
        });
    } catch (e: any) {
        alertStore.error(e);
    }
});

async function deleteShop() {
    try {
        await trpcClient.organization.shop.delete.mutate({ orgId, shopId });

        router.push('/account/shops');
    } catch (error: any) {
        alertStore.error(error);
    }
}
</script>
    
<template>
    <header-container
        :title="'Edit ' + shop.name"
    >
        <router-link
            :to="{ 
                name: 'account.shops.detail',
                params: { 
                    organizationId: orgId,
                    shopId: shopId 
                } 
            }"
            type="button"
            class="group btn"
        >
            Cancel
        </router-link>
    </header-container>
    <main-container v-if="authStore.user">
        <form
            @submit="onSubmit"
        >
            <form-group
                title="Shop information"
            >
                <div class="sm:col-span-6">
                    <label
                        for="Name"
                        class="block text-sm font-medium mb-1"
                    >
                        Name
                    </label>
                    <input
                        id="name"
                        v-model="values.name"
                        type="text"
                        name="name"
                        autocomplete="name"
                        class="field"
                        :class="{ 'is-invalid': errors.name }"
                    >
                    <div class="text-red-700">
                        {{ errors.name }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label
                        for="newOrgId"
                        class="block text-sm font-medium mb-1"
                    >
                        Organization
                    </label>
                    <select
                        id="newOrgId"
                        name="newOrgId"
                        class="field"
                        :value="shop.organizationId"
                        @change="values.newOrgId = parseInt(($event.target! as HTMLSelectElement).value)"
                    >
                        <option
                            v-for="organization in organizations"
                            :key="organization.id"
                            :value="organization.id"
                        >
                            {{ organization.name }}
                        </option>
                    </select>
                    <div class="text-red-700">
                        {{ errors.newOrgId }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label
                        for="shopUrl"
                        class="block text-sm font-medium mb-1"
                    >
                        URL
                    </label>
                    <input
                        id="shopUrl"
                        v-model="values.shopUrl"
                        type="text"
                        name="shopUrl"
                        autocomplete="url"
                        class="field"
                        :class="{ 'is-invalid': errors.shopUrl }"
                    >
                    <div class="text-red-700">
                        {{ errors.shopUrl }}
                    </div>
                </div>
            </form-group>

            <form-group
                title="Integration"
            >
                <template #info>
                    <p>
                        The created integration must have access to following 
                        <a href="https://github.com/FriendsOfShopware/shopmon/blob/main/app/manifest.xml#L18">permissions</a>
                    </p>
                </template>
                <template #default>
                    <div class="sm:col-span-6">
                        <label
                            for="clientId"
                            class="block text-sm font-medium mb-1"
                        > Client-ID </label>
                        <Field
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
                        >
                            Client-Secret
                        </label>
                        <Field
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
                </template>
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
        </Form>

        <form-group :title="'Deleting shop ' + shop.name">
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
    