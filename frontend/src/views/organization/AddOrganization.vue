<template>
    <header-container title="New Organization" />
    <main-container v-if="session.data?.user">
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="owner"
            @submit="onCreateOrganization"
        >
            <form-group
                title="Organization Information"
                sub-title=""
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
                    />
                    <div class="field-error-message">
                        {{ errors.name }}
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
import { trpcClient } from '@/helpers/trpc';

import { Field, Form as VeeForm } from 'vee-validate';
import { useRouter } from 'vue-router';
import * as Yup from 'yup';

const session = authClient.useSession();

const { error } = useAlert();
const router = useRouter();

const schema = Yup.object().shape({
    name: Yup.string().required('Name for organization is required'),
});

const owner = {
    ownerId: session.value.data?.user?.id,
};

async function onCreateOrganization(values: Yup.InferType<typeof schema>) {
    try {
        await trpcClient.organization.create.mutate(values.name);
        await router.push({ name: 'account.organizations.list' });
    } catch (e) {
        error(e instanceof Error ? e.message : String(e));
    }
}
</script>
