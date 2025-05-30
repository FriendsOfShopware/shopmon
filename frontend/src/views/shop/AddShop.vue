<template>
    <header-container title="New Shop" />
    <main-container>
        <vee-form
            v-if="organizations"
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="shops"
            @submit="onSubmit"
        >
            <form-group title="Shop information">
                <div>
                    <label for="Name">Name</label>

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

                <div>
                    <label for="orgId">Organization</label>

                    <field
                        id="orgId"
                        as="select"
                        name="orgId"
                        class="field"
                    >
                        <option
                            v-for="organization in organizations"
                            :key="organization.id"
                            :value="organization.id"
                        >
                            {{ organization.name }}
                        </option>
                    </field>

                    <div class="field-error-message">
                        {{ errors.orgId }}
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
                        :class="{ 'has-error': errors.shop_url }"
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
    </main-container>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { useShopStore } from '@/stores/shop.store';

import {
    type RouterInput,
    type RouterOutput,
    trpcClient,
} from '@/helpers/trpc';
import { Field, Form as VeeForm } from 'vee-validate';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import * as Yup from 'yup';

const shopStore = useShopStore();
const { error } = useAlert();
const router = useRouter();

const organizations = ref<RouterOutput['account']['listOrganizations']>();

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
    shopUrl: Yup.string()
        .required('Shop URL is required')
        .test('is-url-valid', 'Shop URL is not valid', (value) =>
            isValidUrl(value),
        ),
    orgId: Yup.number().required('Organization is required'),
    clientId: Yup.string().required('Client ID is required'),
    clientSecret: Yup.string().required('Client Secret is required'),
});

const shops = {
    orgId: organizations.value?.[0].id,
};

trpcClient.account.listOrganizations.query().then((data) => {
    shops.orgId = data?.[0]?.id;
    organizations.value = data;
});

async function onSubmit(values: RouterInput['organization']['shop']['create']) {
    try {
        await shopStore.createShop(values);

        router.push({ name: 'account.shops.list' });
    } catch (e) {
        error(e instanceof Error ? e.message : String(e));
    }
}
</script>
