<template>
  <router-view />
</template>

<script setup lang="ts">
import { useHead } from "@unhead/vue";
import { computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

import { useDarkMode } from "./composables/useDarkMode";
import { useLocale } from "./composables/useLocale";
import { api, getToken } from "./helpers/api";

const DEFAULT_TITLE = "Shopmon";
const route = useRoute();
const { t } = useI18n();
const { locale, setLocale } = useLocale();

useHead({
  title: computed(() => {
    const titleKey = route.meta.titleKey;
    return typeof titleKey === "string" ? t(titleKey) : undefined;
  }),
  titleTemplate: (title) => (title ? `${title} | ${DEFAULT_TITLE}` : DEFAULT_TITLE),
});

useDarkMode();

// Reconcile the locale with the server on startup. The device/browser
// preference wins when it differs (and is pushed to the server, so users who
// chose a language before this feature keep it); otherwise we adopt the stored
// server preference. Keeps the UI language and notification-email language in
// sync across devices.
onMounted(async () => {
  if (!getToken()) {
    return;
  }
  const { data } = await api.GET("/account/me");
  if (!data) {
    return;
  }
  const current = String(locale.value);
  if (current && current !== data.locale) {
    setLocale(current, true);
  } else {
    setLocale(data.locale, false);
  }
});
</script>
