<template>
  <div class="flex items-center justify-center min-h-[50vh]">
    <div class="text-center text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-8 mx-auto mb-2" />
      <p>{{ action === 'accept' ? $t('organization.acceptingInvitation') : $t('organization.rejectingInvitation') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
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
  api
    .POST("/auth/invitations/{invitationId}/accept", {
      params: { path: { invitationId: route.params.token as string } },
    })
    .then(({ error: respError }) => {
      if (respError) {
        error(
          (respError as { message?: string }).message ?? t("organization.failedAcceptInvitation"),
        );
        router.push({ name: "account.organizations.list" });
      } else {
        router.push({ name: "account.organizations.list" });
        success(t("organization.invitationAccepted"));
      }
    });
} else {
  api
    .POST("/auth/invitations/{invitationId}/reject", {
      params: { path: { invitationId: route.params.token as string } },
    })
    .then(({ error: respError }) => {
      if (respError) {
        error(
          (respError as { message?: string }).message ?? t("organization.failedRejectInvitation"),
        );
        router.push({ name: "account.organizations.list" });
      } else {
        router.push({ name: "account.organizations.list" });
        success(t("organization.invitationRejected"));
      }
    });
}
</script>
