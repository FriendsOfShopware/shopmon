<template>
  <footer class="app-footer">
    <div class="container">
      <div class="footer-links">
        <button v-if="session" type="button" class="footer-link" @click="open">
          {{ $t("footer.whatsNew") }}
        </button>
        <span v-if="session" class="footer-separator">|</span>
        <router-link :to="{ name: 'privacy' }" class="footer-link">
          {{ $t("footer.privacy") }}
        </router-link>
        <span class="footer-separator">|</span>
        <router-link :to="{ name: 'imprint' }" class="footer-link">
          {{ $t("footer.legalNotice") }}
        </router-link>
      </div>
    </div>

    <WhatsNewModal :show="showWhatsNew" @close="dismiss" />
  </footer>
</template>

<script setup lang="ts">
import WhatsNewModal from "@/components/modal/WhatsNewModal.vue";
import { useWhatsNew } from "@/composables/useWhatsNew";
import { useSession } from "@/composables/useSession";
import { computed } from "vue";

const { session } = useSession();
const { isOpen, open, dismiss } = useWhatsNew();

const showWhatsNew = computed(() => {
  return !!session.value && isOpen.value;
});
</script>

<style scoped>
.app-footer {
  margin-top: auto;
  padding: 2rem 0;
  border-top: 1px solid var(--panel-border-color);
  background-color: var(--panel-background);
}

.footer-links {
  text-align: center;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
}

.footer-link {
  color: var(--text-color-muted);
  text-decoration: none;
  font-size: 0.875rem;
  transition: color 0.2s ease;
  border: 0;
  padding: 0;
  background: transparent;
}

.footer-link:hover {
  color: var(--primary-color);
}

.footer-separator {
  color: var(--text-color-muted);
  font-size: 0.875rem;
}
</style>
