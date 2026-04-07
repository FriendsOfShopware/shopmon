<template>
  <footer class="mt-auto border-t">
    <div
      class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-6 py-4 max-sm:flex-col max-sm:text-center"
    >
      <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Logo class="size-4" />
        <span>Shopmon</span>
        <span class="text-border">·</span>
        <span>{{ $t("footer.tagline") }}</span>
      </div>

      <nav class="flex items-center gap-4 text-xs">
        <button
          v-if="session"
          type="button"
          class="text-muted-foreground transition-colors hover:text-foreground"
          @click="open"
        >
          {{ $t("footer.whatsNew") }}
        </button>
        <RouterLink
          :to="{ name: 'privacy' }"
          class="text-muted-foreground transition-colors hover:text-foreground"
        >
          {{ $t("footer.privacy") }}
        </RouterLink>
        <RouterLink
          :to="{ name: 'imprint' }"
          class="text-muted-foreground transition-colors hover:text-foreground"
        >
          {{ $t("footer.legalNotice") }}
        </RouterLink>
        <a
          href="https://github.com/FriendsOfShopware/shopmon/"
          target="_blank"
          rel="noopener"
          class="text-muted-foreground transition-colors hover:text-foreground"
        >
          <icon-mdi:github class="size-4" />
        </a>
      </nav>
    </div>

    <WhatsNewModal :show="showWhatsNew" @close="dismiss" />
  </footer>
</template>

<script setup lang="ts">
import WhatsNewModal from "@/components/modal/WhatsNewModal.vue";
import Logo from "@/components/Logo.vue";
import { useWhatsNew } from "@/composables/useWhatsNew";
import { useSession } from "@/composables/useSession";
import { computed } from "vue";

const { session } = useSession();
const { isOpen, open, dismiss } = useWhatsNew();

const showWhatsNew = computed(() => {
  return !!session.value && isOpen.value;
});
</script>
