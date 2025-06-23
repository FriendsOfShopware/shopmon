<template>
    <header-container title="New Organization" />
    <main-container v-if="session.data?.user">
        <vee-form
            v-slot="{ errors, isSubmitting, setFieldValue }"
            :validation-schema="schema"
            class="panel"
            @submit="onCreateOrganization"
        >
            <form-group
                title="Organization Information"
            >
                <div>
                    <label for="Name">Name</label>
                    <field
                        id="name"
                        type="text"
                        name="name"
                        autocomplete="name"
                        class="field"
                        :class="{ 'has-error': errors.name }"
                        @input="(e) => onNameChange(e, setFieldValue)"
                    />
                    <div class="field-error-message">
                        {{ errors.name }}
                    </div>
                </div>

                <div>
                    <label for="slug">Slug</label>
                    <field
                        id="slug"
                        type="text"
                        name="slug"
                        autocomplete="slug"
                        class="field"
                        :class="{ 'has-error': errors.slug }"
                        @input="onSlugManualEdit"
                    />
                    <div class="field-error-message">
                        {{ errors.slug }}
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

import { Field, Form as VeeForm } from 'vee-validate';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import * as Yup from 'yup';

const session = authClient.useSession();

const { error } = useAlert();
const router = useRouter();

// Track if slug has been manually edited
const slugManuallyEdited = ref(false);

const schema = Yup.object().shape({
    name: Yup.string().required('Name for organization is required'),
    slug: Yup.string()
        .required('Slug for organization is required')
        .matches(
            /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
            'Slug must be lowercase and can only contain letters, numbers, and hyphens',
        ),
});

function generateSlug(str: string): string {
    return str
        .toLowerCase()
        .trim()
        .replace(/[^\w\s-]/g, '') // Remove special characters
        .replace(/[\s_-]+/g, '-') // Replace spaces and underscores with hyphens
        .replace(/^-+|-+$/g, ''); // Remove leading/trailing hyphens
}

function onNameChange(
    event: Event,
    // eslint-disable-next-line no-unused-vars
    setFieldValue: (field: string, value: unknown) => void,
) {
    if (!slugManuallyEdited.value) {
        const target = event.target as HTMLInputElement;
        const slug = generateSlug(target.value);
        setFieldValue('slug', slug);
    }
}

function onSlugManualEdit() {
    slugManuallyEdited.value = true;
}

async function onCreateOrganization(values: Record<string, unknown>) {
    const typedValues = values as Yup.InferType<typeof schema>;
    try {
        const resp = await authClient.organization.create({
            name: typedValues.name,
            slug: typedValues.slug,
        });
        if (resp.error) {
            error(resp.error.message ?? 'Failed to create organization');
            return;
        }
        await router.push({ name: 'account.organizations.list' });
    } catch (e) {
        error(e instanceof Error ? e.message : String(e));
    }
}
</script>
