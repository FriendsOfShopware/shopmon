<script setup lang="ts">
import { watch } from "vue";
import { useRoute } from "vue-router";
import { useSidebar } from "@/components/ui/sidebar";

const route = useRoute();
const { isMobile, openMobile, setOpenMobile } = useSidebar();

// Close the mobile sidebar sheet whenever the route changes, otherwise it
// stays open over the freshly navigated page. See #659.
watch(
  () => route.fullPath,
  () => {
    if (isMobile.value && openMobile.value) {
      setOpenMobile(false);
    }
  },
);
</script>

<template></template>
