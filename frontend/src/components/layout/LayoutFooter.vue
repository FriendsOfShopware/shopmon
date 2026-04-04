<template>
  <footer class="mt-auto border-t bg-card py-8">
    <div class="container mx-auto px-4">
      <div class="flex items-center justify-center gap-4 text-sm text-muted-foreground">
        <button v-if="session" type="button" class="hover:text-primary transition-colors bg-transparent border-0 p-0" @click="open">
          {{ $t("footer.whatsNew") }}
        </button>
        <span v-if="session">|</span>
        <router-link :to="{ name: 'privacy' }" class="hover:text-primary transition-colors">
          {{ $t("footer.privacy") }}
        </router-link>
        <span>|</span>
        <router-link :to="{ name: 'imprint' }" class="hover:text-primary transition-colors">
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
