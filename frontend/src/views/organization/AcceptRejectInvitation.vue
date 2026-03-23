<template>YAYYY</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { useRoute, useRouter } from "vue-router";

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
        error((respError as { message?: string }).message ?? "Failed to accept invitation");
        router.push({ name: "account.organizations.list" });
      } else {
        router.push({ name: "account.organizations.list" });
        success("Invitation accepted successfully!");
      }
    });
} else {
  api
    .POST("/auth/invitations/{invitationId}/reject", {
      params: { path: { invitationId: route.params.token as string } },
    })
    .then(({ error: respError }) => {
      if (respError) {
        error((respError as { message?: string }).message ?? "Failed to reject invitation");
        router.push({ name: "account.organizations.list" });
      } else {
        router.push({ name: "account.organizations.list" });
        success("Invitation rejected successfully!");
      }
    });
}
</script>
