<template>
  <div v-if="isImpersonating" class="sticky top-0 z-10 bg-amber-500 text-white shadow">
    <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-2 max-sm:flex-col max-sm:text-center">
      <div class="flex items-center gap-2 text-sm">
        <icon-fa6-solid:user-secret class="size-4" />
        <span v-html="$t('impersonation.banner', { email: session?.user?.email })" />
      </div>
      <Button size="sm" variant="outline" class="border-white/30 bg-white/20 text-white hover:bg-white/30" @click="stopImpersonating">
        {{ $t("impersonation.stop") }}
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import { computed } from "vue";
import { Button } from "@/components/ui/button";

const { t } = useI18n();
const { session } = useSession();
const alert = useAlert();

const isImpersonating = computed(() => {
  if (!session.value) return false;
  return !!session.value.session?.impersonatedBy;
});

async function stopImpersonating() {
  try {
    const { error } = await api.POST("/auth/admin/stop-impersonating");

    if (error) {
      throw new Error("Failed to stop impersonating");
    }

    window.location.reload();
  } catch (error) {
    alert.error(
      t("impersonation.failedStop", {
        error: error instanceof Error ? error.message : "Unknown error",
      }),
    );
  }
}
</script>
