<script setup lang="ts">
    import { storeToRefs } from 'pinia';

    import Header from '@/components/layout/Header.vue';
    import MainContainer from '@/components/layout/MainContainer.vue';
    import FormGroup from '@/components/layout/FormGroup.vue';
    
    import { useAlertStore } from '@/stores/alert.store';
    import { useAuthStore } from '@/stores/auth.store';
    import { useTeamStore } from '@/stores/team.store';
    
    import { Form, Field } from 'vee-validate';
    import { useRouter, useRoute } from 'vue-router';
    import * as Yup from 'yup';
    import { ref } from 'vue';
    
    const authStore = useAuthStore();
    const teamStore = useTeamStore();
    const alertStore = useAlertStore();
    const router = useRouter();
    const route = useRoute();

    const { user } = storeToRefs(authStore);
    
    const teamId = parseInt(route.params.teamId as string, 10);    
    const team = user.value?.teams.find(team => team.id == teamId);
    
    const showTeamDeletionModal = ref(false)
    
    const schema = Yup.object().shape({
        name: Yup.string().required('Team name is required'),
    });

    const schemaMembers = Yup.object().shape({
        email: Yup.string().email('Email address is not valid').required('Email address is required'),
    });
    
    async function onSubmit(values: any) {
        console.log("Waiting for feature: Update Team");
    }

    async function onAddMember(values: any) {
        if ( team ) {
            try {
                await teamStore.addMember(team.id, values);
    
                router.push('/account/teams');
            } catch (error: Error) {
                alertStore.error(error);
            }
        }
    }
    
    async function deleteTeam() {
        console.log('Waiting for feature: Create Team');
        return;
        
        if ( team ) {
            try {
                await teamStore.delete(team.id);
    
                router.push('/account/teams');
            } catch (error: Error) {
                alertStore.error(error);
            }
        }
    }
    
    </script>
        
    <template>
        <Header :title="'Edit ' + team.name" v-if="team">
            <router-link :to="{ name: 'account.teams.detail', params: { teamId: team.id } }"
                type="button" class="group btn">
                Cancel
            </router-link>
        </Header>
        <MainContainer v-if="team && authStore.user">
            <Form 
                v-slot="{ errors, isSubmitting }"
                :validation-schema="schema"
                :initial-values="team"
                @submit="onSubmit"
            >
                <FormGroup title="Team Information" subTitle="">
                    <div class="sm:col-span-6">
                        <label for="Name" class="block text-sm font-medium text-gray-700 mb-1"> Name </label>
                        <Field type="text" name="name" id="name" autocomplete="name" class="field"
                            v-bind:class="{ 'is-invalid': errors.name }" />
                        <div class="text-red-700">
                            {{ errors.name }}
                        </div>
                    </div>
                </FormGroup>

                <div class="flex justify-end pb-12">
                    <button disabled type="submit" class="btn btn-primary opacity-25">
                        <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                            <icon-fa6-solid:floppy-disk class="h-5 w-5" aria-hidden="true"
                                v-if="!isSubmitting" />
                            <icon-line-md:loading-twotone-loop class="w-5 h-5" v-else />
                        </span>
                        Save
                    </button>
                </div>
            </Form>

            <Form 
                v-slot="{ errors, isSubmitting }"
                :validation-schema="schemaMembers"
                @submit="onAddMember"
            >
                <FormGroup title="Members" subTitle="">
                    <div class="sm:col-span-6">
                        <label for="email" class="block text-sm font-medium text-gray-700 mb-1"> Email </label>
                        <Field type="text" name="email" id="email" autocomplete="email" class="field"
                            v-bind:class="{ 'is-invalid': errors.emailAddress }" />
                        <div class="text-red-700">
                            {{ errors.email }}
                        </div>
                    </div>
                </FormGroup>

                <div class="flex justify-end pb-12">
                    <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
                        <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                            <icon-fa6-solid:plus class="h-5 w-5" aria-hidden="true"
                                v-if="!isSubmitting" />
                            <icon-line-md:loading-twotone-loop class="w-5 h-5" v-else />
                        </span>
                        Add
                    </button>
                </div>
            </Form>
    
            <FormGroup :title="'Deleting team ' + team.name">
                <form action="#" method="POST">

                    <p>Once you delete your team, you will lose all data associated with it. </p>
    
                    <div class="mt-5">
                        <button disabled type="button" class="btn btn-danger group flex items-center opacity-25"
                            @click="showTeamDeletionModal = true">
                            <icon-fa6-solid:trash
                                class="w-4 h-4 -ml-1 mr-2 opacity-25 group-hover:opacity-50" />
                            Delete team
                        </button>
                    </div>
                </form>
            </FormGroup>
    
            <Modal :show="showTeamDeletionModal" @close="showTeamDeletionModal = false">
                <template #icon>
                    <icon-fa6-solid:triangle-exclamation class="h-6 w-6 text-red-600" aria-hidden="true" />
                </template>
                <template #title>Delete team</template>
                <template #content>
                    Are you sure you want to delete your Team? All of your data will
                    be permanently removed
                    from our servers forever. This action cannot be undone.
                </template>
                <template #footer>
                    <button type="button" class="btn btn-danger w-full sm:ml-3 sm:w-auto"
                        @click="deleteTeam">
                        Delete
                    </button>
                    <button ref="cancelButtonRef" type="button"
                        class="btn w-full mt-3 sm:w-auto sm:mt-0"
                        @click="showTeamDeletionModal = false">
                        Cancel
                    </button>
                </template>
            </Modal>
    
        </MainContainer>
    </template>
        