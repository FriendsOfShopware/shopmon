<script setup lang="ts">
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';
import { useRouter } from 'vue-router';

import { trpcClient } from '@/helpers/trpc';
import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';

import { useForm } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import * as z from 'zod';

const authStore = useAuthStore();
const alertStore = useAlertStore();
const router = useRouter();

const organizations = await trpcClient.account.currentUserOrganizations.query();

const { values, handleSubmit, errors, isSubmitting  } = useForm({
    validationSchema: toTypedSchema(z.object({
        name: z.string().min(1, 'Shop name is required'),
        shopUrl: z.string().url('Shop URL is required'),
        orgId: z.number().min(1, 'Team is required'),
        clientId: z.string().min(1, 'Client ID is required'),
        clientSecret: z.string().min(1, 'Client Secret is required'),
    })),
    initialValues: {
        orgId: organizations[0]?.id,
    },
});

const submit = handleSubmit(async (values) => {
    try {
        values.shopUrl = values.shopUrl.replace(/\/+$/, '');
        await trpcClient.organization.shop.create.mutate(values);

        router.push('/account/shops');
    } catch (e: any) {
        alertStore.error(e);
    }
});

</script>

<template>
    <header-container title="New Shop" />
    <main-container v-if="authStore.user">
        <form @submit="submit">
            <form-group
                title="Shop information"
                sub-title=""
            >
                <div class="sm:col-span-6">
                    <label
                        for="Name"
                        class="block text-sm font-medium mb-1"
                    > Name </label>
                    <input
                        id="name"
                        v-model="values.name"
                        type="text"
                        name="name"
                        autocomplete="username"
                        class="field"
                        :class="{ 'is-invalid': errors.name }"
                    >
                    <div class="text-red-700">
                        {{ errors.name }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label
                        for="orgId"
                        class="block text-sm font-medium mb-1"
                    > Team </label>
                    <select
                        id="orgId"
                        v-model="values.orgId"
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
                    </select>
                    <div class="text-red-700">
                        {{ errors.orgId }}
                    </div>
                </div>

                <div class="sm:col-span-6">
                    <label
                        for="shopUrl"
                        class="block text-sm font-medium mb-1"
                    > URL </label>
                    <Field
                        id="shopUrl"
                        type="text"
                        name="shopUrl"
                        autocomplete="url"
                        class="field"
                        :class="{ 'is-invalid': errors.shopUrl }"
                    />
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
                        <input
                            id="clientId"
                            v-model="values.clientId"
                            type="text"
                            name="clientId"
                            class="field"
                            :class="{ 'is-invalid': errors.clientId }"
                        >
                        <div class="text-red-700">
                            {{ errors.clientId }}
                        </div>
                    </div>

                    <div class="sm:col-span-6">
                        <label
                            for="clientSecret"
                            class="block text-sm font-medium mb-1"
                        > Client-Secret </label>
                        <input
                            id="clientSecret"
                            v-model="values.clientSecret"
                            type="text"
                            name="clientSecret"
                            class="field"
                            :class="{ 'is-invalid': errors.clientSecret }"
                        >
                        <div class="text-red-700">
                            {{ errors.clientSecret }}
                        </div>
                    </div>
                </template>
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
        </Form>
    </main-container>
</template>
