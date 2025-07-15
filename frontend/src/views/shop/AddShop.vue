<template>
    <header-container title="New Shop" />
    <main-container>
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="shops"
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
                        class="field"
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
                            required
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
    </main-container>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';

import {
    type RouterInput,
    type RouterOutput,
    trpcClient,
} from '@/helpers/trpc';
import { Field, Form as VeeForm } from 'vee-validate';
import { ref } from 'vue';

import { useRoute, useRouter } from 'vue-router';
import * as Yup from 'yup';

const { error } = useAlert();
const router = useRouter();
const route = useRoute();

const projects = ref<RouterOutput['account']['currentUserProjects']>([]);
const selectedProjectId = ref<number>(
    route.query.projectId ? Number(route.query.projectId) : 0,
);

trpcClient.account.currentUserProjects.query().then((data) => {
    projects.value = data;
    if (!selectedProjectId.value && data.length > 0) {
        selectedProjectId.value = data[0].id;
    }
});

const isValidUrl = (url: string) => {
    try {
        new URL(url);
        // eslint-disable-next-line no-unused-vars
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
    projectId: Yup.number().required('Project is required'),
    clientId: Yup.string().required('Client ID is required'),
    clientSecret: Yup.string().required('Client Secret is required'),
});

const shops = {
    projectId: selectedProjectId.value,
};

const onSubmit = async (values: Record<string, unknown>) => {
    try {
        const typedValues = values as Yup.InferType<typeof schema>;
        const input: RouterInput['organization']['shop']['create'] = {
            name: typedValues.name,
            shopUrl: typedValues.shopUrl.replace(/\/+$/, ''),
            clientId: typedValues.clientId,
            clientSecret: typedValues.clientSecret,
            projectId: selectedProjectId.value,
        };
        await trpcClient.organization.shop.create.mutate(input);

        router.push({ name: 'account.project.list' });
    } catch (e) {
        error(e instanceof Error ? e.message : String(e));
    }
};
</script>
