<script setup lang="ts">
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';

import TextField from '@/components/fields/TextField.vue';

import { trpcClient } from '@/helpers/trpc';
import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';

import { useRouter } from 'vue-router';

import { useForm } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import * as z from 'zod';

const authStore = useAuthStore();
const alertStore = useAlertStore();
const router = useRouter();

const { handleSubmit, isSubmitting  } = useForm({
    validationSchema: toTypedSchema(z.object({
        name: z.string().min(1, 'Team name is required'),
    })),
});

const submit = handleSubmit(async (values) => {
    try {
        await trpcClient.organization.create.mutate(values.name);
        await router.push({ name: 'account.organizations.list' });
    } catch (e: any) {
        alertStore.error(e);
    }
});
</script>

<template>
    <header-container title="New Organization" />
    <main-container v-if="authStore.user">
        <form
            @submit="submit"
        >
            <form-group
                title="Team Information"
                sub-title=""
            >
                <div class="sm:col-span-6">
                    <text-field
                        label="Name"
                        name="name"
                        autocomplete="name"
                    />
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
        </Form>
    </main-container>
</template>
