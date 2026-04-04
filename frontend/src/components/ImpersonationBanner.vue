<template>
  <div v-if="isImpersonating" class="impersonation-banner">
    <div class="impersonation-content">
      <div class="impersonation-text">
        <icon-fa6-solid:user-secret class="impersonation-icon" />
        <span v-html="$t('impersonation.banner', { email: session?.user?.email })" />
      </div>
      <button class="btn btn-sm btn-stop" @click="stopImpersonating">
        {{ $t("impersonation.stop") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import { computed } from "vue";

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

    // Force a complete page reload to ensure clean session state
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

<style scoped>
.impersonation-banner {
  background-color: #f59e0b;
  color: #ffffff;
  padding: 0.75rem 0;
  position: sticky;
  top: 0;
  z-index: 10;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.impersonation-content {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.impersonation-text {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.impersonation-icon {
  width: 1rem;
  height: 1rem;
}

.btn-stop {
  background-color: rgba(255, 255, 255, 0.2);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 0.25rem 0.75rem;
  font-size: 0.875rem;
}

.btn-stop:hover {
  background-color: rgba(255, 255, 255, 0.3);
  border-color: rgba(255, 255, 255, 0.4);
}

@media (max-width: 640px) {
  .impersonation-content {
    flex-direction: column;
    text-align: center;
  }
}
</style>
