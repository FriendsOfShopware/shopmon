<template>
    <header-container title="New Shop" />
    <main-container v-if="authStore.user">
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="shops"
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
                        class="field"
                        :class="{ 'is-invalid': errors.name }"
                    />
                    <div class="text-red-700">
                        {{ errors.name }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label
                        for="orgId"
                        class="block text-sm font-medium mb-1"
                    > Organization </label>
                    <field
                        id="orgId"
                        as="select"
                        name="orgId"
                        class="field"
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
                        {{ errors.orgId }}
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
                        :class="{ 'is-invalid': errors.shop_url }"
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
                    > Client-Secret </label>
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

            <div class="flex justify-end group">
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
    </main-container>
</template>

<script setup lang="ts">
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';

import { Form as VeeForm, Field } from 'vee-validate';
import { useRouter } from 'vue-router';
import * as Yup from 'yup';
import { RouterInput } from '@/helpers/trpc';

const authStore = useAuthStore();
const shopStore = useShopStore();
const alertStore = useAlertStore();
const router = useRouter();

const isValidUrl = (url: string) => {
    try {
        new URL(url);
    } catch (e) {
        return false;
    }
    return true;
};

const schema = Yup.object().shape({
    name: Yup.string().required('Shop name is required'),
    shopUrl: Yup.string().required('Shop URL is required')
        .test('is-url-valid', 'Shop URL is not valid', (value) => isValidUrl(value)),
    orgId: Yup.number().required('Organization is required'),
    clientId: Yup.string().required('Client ID is required'),
    clientSecret: Yup.string().required('Client Secret is required'),
});

const shops = {
    orgId: authStore.user?.organizations[0]?.id,
};

async function onSubmit(values: RouterInput['organization']['shop']['create']) {
    try {
        await shopStore.createShop(values);

        router.push('/account/shops');
    } catch (e: any) {
        alertStore.error(e);
    }
}
</script>
