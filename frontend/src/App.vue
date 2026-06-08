<template>
  <router-view />
</template>

<script setup lang="ts">
import { useHead } from "@unhead/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

import { useDarkMode } from "./composables/useDarkMode";

const DEFAULT_TITLE = "Shopmon";
const route = useRoute();
const { t } = useI18n();

useHead({
  title: computed(() => {
    const titleKey = route.meta.titleKey;
    return typeof titleKey === "string" ? t(titleKey) : undefined;
  }),
  titleTemplate: (title) => (title ? `${title} | ${DEFAULT_TITLE}` : DEFAULT_TITLE),
});

useDarkMode();
</script>
