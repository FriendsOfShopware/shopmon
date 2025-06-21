<template>
    <header-container title="New Project" />
    <main-container>
        <vee-form
            v-if="!organizations.isPending"
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="initialValues"
            class="panel"
            @submit="onSubmit"
        >
            <form-group title="Project information">
                <div>
                    <label for="name">Name</label>

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
                    <label for="description">Description</label>

                    <field
                        id="description"
                        v-slot="{ field }"
                        name="description"
                    >
                        <textarea
                            v-bind="field"
                            id="description"
                            class="field"
                            rows="4"
                            placeholder="Optional project description..."
                            :class="{ 'has-error': errors.description }"
                        />
                    </field>

                    <div class="field-error-message">
                        {{ errors.description }}
                    </div>
                </div>

                <div>
                    <label for="organizationId">Organization</label>

                    <field
                        id="organizationId"
                        name="organizationId"
                    >
                        <select
                            v-model="selectedOrgId"
                            class="field"
                            :class="{ 'has-error': errors.organizationId }"
                        >
                            <option value="">Select an organization</option>
                            <option
                                v-for="organization in organizations.data"
                                :key="organization.id"
                                :value="organization.id"
                            >
                                {{ organization.name }}
                            </option>
                        </select>
                    </field>

                    <div class="field-error-message">
                        {{ errors.organizationId }}
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
import { authClient } from '@/helpers/auth-client';
import { type RouterInput, trpcClient } from '@/helpers/trpc';
import { Field, Form as VeeForm } from 'vee-validate';
import { ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import * as Yup from 'yup';

const { error } = useAlert();
const router = useRouter();

const organizations = authClient.useListOrganizations();
const selectedOrgId = ref<string>('');

const schema = Yup.object().shape({
    name: Yup.string().required('Project name is required'),
    description: Yup.string().optional(),
    organizationId: Yup.string().required('Organization is required'),
});

const initialValues = {
    name: '',
    description: '',
    organizationId: organizations.value.data?.[0]?.id ?? '',
};

watch(
    organizations,
    (newValue) => {
        if (newValue.data?.length && !selectedOrgId.value) {
            selectedOrgId.value = newValue.data[0].id;
            initialValues.organizationId = newValue.data[0].id;
        }
    },
    { immediate: true },
);

const onSubmit = async (values: Record<string, unknown>) => {
    try {
        const typedValues = values as Yup.InferType<typeof schema>;
        const input: RouterInput['organization']['project']['create'] = {
            orgId: typedValues.organizationId,
            name: typedValues.name,
            description: typedValues.description ?? undefined,
        };
        await trpcClient.organization.project.create.mutate(input);

        router.push({ name: 'account.project.list' });
    } catch (e) {
        error(e instanceof Error ? e.message : String(e));
    }
};
</script>
