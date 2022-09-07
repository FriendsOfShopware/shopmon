<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';
import Spinner from '@/components/icon/Spinner.vue';

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';
import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';

import { Form, Field } from 'vee-validate';
import { useRouter, useRoute } from 'vue-router';
import * as Yup from 'yup';
import { ref } from 'vue';

const authStore = useAuthStore();
const shopStore = useShopStore();
const alertStore = useAlertStore();
const router = useRouter();
const route = useRoute();

shopStore.loadShop(route.params.teamId as string, route.params.shopId as string);

const showShopDeletionModal = ref(false)

const shop = {
    id: shopStore.shop?.id,
    name: shopStore.shop?.name,
    teamId: shopStore.shop?.team_id,
    shop_url: shopStore.shop?.url
}

const schema = Yup.object().shape({
    name: Yup.string().required('Shop name is required'),
    shop_url: Yup.string().required('Shop URL is required').url(),
    teamId: Yup.number().required('Team is required'),
});

async function onSubmit(values: any) {
    if ( shop.teamId && shop.id ) {
        try {
            await shopStore.updateShop(shop.teamId, shop.id, values);

            router.push({
                name: 'account.shops.detail',
                params: {
                    teamId: shop.teamId,
                    shopId: shop.id
                }
            })
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}

async function deleteShop() {
    if ( shop.teamId && shop.id ) {
        try {
            await shopStore.delete(shop.teamId, shop.id);

            router.push('/account/shops');
        } catch (error: Error) {
            alertStore.error(error);
        }
    }
}

</script>
    
<template>
    <Header :title="'Edit ' + shop.name" v-if="shop">
        <router-link :to="{ name: 'account.shops.detail', params: { teamId: shop.teamId, shopId: shop.id } }"
            type="button" class="group btn">
            Cancel
        </router-link>
    </Header>
    <MainContainer v-if="shop && authStore.user">
        <Form v-slot="{ errors, isSubmitting }" :validation-schema="schema" :initial-values="shop" @submit="onSubmit">
            <FormGroup title="Shop information" subTitle="">
                <div class="sm:col-span-6">
                    <label for="Name" class="block text-sm font-medium text-gray-700 mb-1"> Name </label>
                    <Field type="text" name="name" id="name" autocomplete="name" class="field"
                        v-bind:class="{ 'is-invalid': errors.name }" />
                    <div class="text-red-700">
                        {{ errors.name }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label for="teamId" class="block text-sm font-medium text-gray-700 mb-1"> Team </label>
                    <Field as="select" id="teamId" name="teamId" class="field">
                        <option v-for="team in authStore.user.teams" :value="team.id" :key="team.id">
                            {{ team.name }}
                        </option>
                    </Field>
                    <div class="text-red-700">
                        {{ errors.teamId }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label for="shop_url" class="block text-sm font-medium text-gray-700 mb-1"> URL </label>
                    <Field type="text" name="shop_url" id="shop_url" autocomplete="url" class="field"
                        v-bind:class="{ 'is-invalid': errors.shop_url }" />
                    <div class="text-red-700">
                        {{ errors.shop_url }}
                    </div>
                </div>

            </FormGroup>

            <FormGroup title="Integration"
                info="<p>The created integration must have access to following <a href='https://github.com/FriendsOfShopware/shopmon/blob/main/app/manifest.xml#L18'>permissions</a></p>">
                <div class="sm:col-span-6">
                    <label for="client_id" class="block text-sm font-medium text-gray-700 mb-1"> Client-ID </label>
                    <Field type="text" name="client_id" id="client_id" class="field"
                        v-bind:class="{ 'is-invalid': errors.client_id }" />
                    <div class="text-red-700">
                        {{ errors.client_id }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label for="client_secret" class="block text-sm font-medium text-gray-700 mb-1"> Client-Secret
                    </label>
                    <Field type="text" name="client_secret" id="client_secret" class="field"
                        v-bind:class="{ 'is-invalid': errors.client_secret }" />
                    <div class="text-red-700">
                        {{ errors.client_secret }}
                    </div>
                </div>
            </FormGroup>

            <div class="flex justify-end pb-12">
                <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
                    <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                        <font-awesome-icon class="h-5 w-5" aria-hidden="true" icon="fa-solid fa-floppy-disk"
                            v-if="!isSubmitting" />
                        <Spinner v-else />
                    </span>
                    Save
                </button>
            </div>
        </Form>

        <FormGroup :title="'Deleting shop ' + shop.name">
            <form action="#" method="POST">

                <p>Once you delete your shop, you will lose all data associated with it. </p>

0.                <div class="mt-5">
                    <button type="button" class="btn btn-danger group flex items-center"
                        @click="showShopDeletionModal = true">
                        <font-awesome-icon icon="fa-solid fa-trash"
                            class="w-4 h-4 -ml-1 mr-2 opacity-25 group-hover:opacity-50" />
                        Delete shop
                    </button>
                </div>
            </form>
        </FormGroup>

        <TransitionRoot as="template" :show="showShopDeletionModal">
            <Dialog as="div" class="relative z-10" @close="showShopDeletionModal = false">
                <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0"
                    enter-to="opacity-100" leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
                    <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
                </TransitionChild>

                <div class="fixed z-10 inset-0 overflow-y-auto">
                    <div class="flex items-end sm:items-center justify-center min-h-full p-4 text-center sm:p-0">
                        <TransitionChild as="template" enter="ease-out duration-300"
                            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                            enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200"
                            leave-from="opacity-100 translate-y-0 sm:scale-100"
                            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
                            <DialogPanel
                                class="relative bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:max-w-lg sm:w-full sm:p-6">
                                <div class="sm:flex sm:items-start">
                                    <font-awesome-icon icon="fa-solid fa-triangle-exclamation"
                                        class="h-6 w-6 text-red-600" aria-hidden="true" />
                                    <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                                        <DialogTitle as="h3" class="text-lg leading-6 font-medium text-gray-900">
                                            Delete shop
                                        </DialogTitle>
                                        <div class="mt-2">
                                            <p class="text-sm text-gray-500">
                                                Are you sure you want to delete your Shop? All of your data will
                                                be permanently removed
                                                from our servers forever. This action cannot be undone.
                                            </p>
                                        </div>
                                    </div>
                                </div>
                                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                    <button type="button" class="btn btn-danger w-full sm:ml-3 sm:w-auto"
                                        @click="deleteShop">
                                        Delete
                                    </button>
                                    <button ref="cancelButtonRef" type="button"
                                        class="btn w-full mt-3 sm:w-auto sm:mt-0"
                                        @click="showShopDeletionModal = false">
                                        Cancel
                                    </button>
                                </div>
                            </DialogPanel>
                        </TransitionChild>
                    </div>
                </div>
            </Dialog>
        </TransitionRoot>

    </MainContainer>
</template>
    