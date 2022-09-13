<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';

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

const teamId = parseInt(route.params.teamId as string, 10);
const shopId = parseInt(route.params.shopId as string, 10);

shopStore.loadShop(teamId, shopId);

const showShopDeletionModal = ref(false)

const schema = Yup.object().shape({
    name: Yup.string().required('Shop name is required'),
    url: Yup.string().required('Shop URL is required').url(),
    team_id: Yup.number().required('Team is required'),
});

async function onSubmit(values: any) {
    if ( shopStore.shop ) {
        try {
            await shopStore.updateShop(shopStore.shop.team_id, shopStore.shop.id, values);

            router.push({
                name: 'account.shops.detail',
                params: {
                    teamId: shopStore.shop.team_id,
                    shopId: shopStore.shop.id
                }
            })
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}

async function deleteShop() {
    if ( shopStore.shop ) {
        try {
            await shopStore.delete(shopStore.shop.team_id, shopStore.shop.id);

            router.push('/account/shops');
        } catch (error: Error) {
            alertStore.error(error);
        }
    }
}

</script>
    
<template>
    <Header :title="'Edit ' + shopStore.shop.name" v-if="shopStore.shop">
        <router-link :to="{ name: 'account.shops.detail', params: { teamId: shopStore.shop.team_id, shopId: shopStore.shop.id } }"
            type="button" class="group btn">
            Cancel
        </router-link>
    </Header>
    <MainContainer v-if="shopStore.shop && authStore.user">
        <Form v-slot="{ errors, isSubmitting }" :validation-schema="schema" :initial-values="shopStore.shop" @submit="onSubmit">
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
                    <label for="team_id" class="block text-sm font-medium text-gray-700 mb-1"> Team </label>
                    <Field as="select" id="team_id" name="team_id" class="field">
                        <option v-for="team in authStore.user.teams" :value="team.id" :key="team.id">
                            {{ team.name }}
                        </option>
                    </Field>
                    <div class="text-red-700">
                        {{ errors.teamId }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label for="url" class="block text-sm font-medium text-gray-700 mb-1"> URL </label>
                    <Field type="text" name="url" id="url" autocomplete="url" class="field"
                        v-bind:class="{ 'is-invalid': errors.url }" />
                    <div class="text-red-700">
                        {{ errors.url }}
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
                        <icon-fa6-solid:floppy-disk class="h-5 w-5" aria-hidden="true"
                            v-if="!isSubmitting" />
                        <icon-line-md:loading-twotone-loop class="w-5 h-5" v-else />
                    </span>
                    Save
                </button>
            </div>
        </Form>

        <FormGroup :title="'Deleting shop ' + shopStore.shop.name">
            <form action="#" method="POST">

                <p>Once you delete your shop, you will lose all data associated with it. </p>

                <div class="mt-5">
                    <button type="button" class="btn btn-danger group flex items-center"
                        @click="showShopDeletionModal = true">
                        <icon-fa6-solid:trash
                            class="w-4 h-4 -ml-1 mr-2 opacity-25 group-hover:opacity-50" />
                        Delete shop
                    </button>
                </div>
            </form>
        </FormGroup>

        <Modal :show="showShopDeletionModal" :closeXMark="true" @close="showShopDeletionModal = false">
            <template #icon><icon-fa6-solid:triangle-exclamation class="h-6 w-6 text-red-600" aria-hidden="true" /></template>
            <template #title>Delete shop</template>
            <template #content>                
                Are you sure you want to delete your Shop? All of your data will
                be permanently removed
                from our servers forever. This action cannot be undone.
            </template>
            <template #footer>
                <button type="button" class="btn btn-danger w-full sm:ml-3 sm:w-auto"
                    @click="deleteShop">
                    Delete
                </button>
                <button ref="cancelButtonRef" type="button"
                    class="btn w-full mt-3 sm:w-auto sm:mt-0"
                    @click="showShopDeletionModal = false">
                    Cancel
                </button>
            </template>
        </Modal>

    </MainContainer>
</template>
    