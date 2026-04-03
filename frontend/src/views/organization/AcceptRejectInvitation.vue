<template>YAYYY</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { authClient } from "@/helpers/auth-client";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

const { t } = useI18n();

const props = defineProps<{
  action: "accept" | "reject";
}>();

const route = useRoute();
const router = useRouter();
const { error, success } = useAlert();

if (props.action === "accept") {
  authClient.organization
    .acceptInvitation({
      invitationId: route.params.token as string,
    })
    .then((resp) => {
      if (resp.error) {
        error(resp.error.message ?? t("organization.failedAcceptInvitation"));
        router.push({ name: "account.organizations.list" });
      } else {
        router.push({ name: "account.organizations.list" });
        success(t("organization.invitationAccepted"));
      }
    });
} else {
  authClient.organization
    .rejectInvitation({
      invitationId: route.params.token as string,
    })
    .then((resp) => {
      if (resp.error) {
        error(resp.error.message ?? t("organization.failedRejectInvitation"));
        router.push({ name: "account.organizations.list" });
      } else {
        router.push({ name: "account.organizations.list" });
        success(t("organization.invitationRejected"));
      }
    });
}
</script>
