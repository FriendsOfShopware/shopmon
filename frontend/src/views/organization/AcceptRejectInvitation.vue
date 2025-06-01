<template>
    YAYYY
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { authClient } from '@/helpers/auth-client';
import { useRoute, useRouter } from 'vue-router';

const props = defineProps<{
    action: 'accept' | 'reject';
}>();

const route = useRoute();
const router = useRouter();
const { error, success } = useAlert();

if (props.action === 'accept') {
    authClient.organization
        .acceptInvitation({
            invitationId: route.params.token as string,
        })
        .then((resp) => {
            if (resp.error) {
                error(resp.error.message || 'Failed to accept invitation');
                router.push({ name: 'account.organizations.list' });
            } else {
                router.push({ name: 'account.organizations.list' });
                success('Invitation accepted successfully!');
            }
        });
} else {
    authClient.organization
        .rejectInvitation({
            invitationId: route.params.token as string,
        })
        .then((resp) => {
            if (resp.error) {
                error(resp.error.message || 'Failed to reject invitation');
                router.push({ name: 'account.organizations.list' });
            } else {
                router.push({ name: 'account.organizations.list' });
                success('Invitation rejected successfully!');
            }
        });
}
</script>
