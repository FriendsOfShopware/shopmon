<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="flex max-h-[calc(100vh-4rem)] max-w-2xl flex-col">
      <DialogHeader>
        <DialogTitle>
          {{ $t("extensionChangelog.title") }} -
          <span class="font-normal text-muted-foreground">{{ extension?.name }}</span>
        </DialogTitle>
      </DialogHeader>

      <ul
        v-if="changelogEntries.length > 0"
        class="-mx-6 min-h-0 list-none space-y-6 overflow-y-auto px-6 py-0"
      >
        <li v-for="changeLog in changelogEntries" :key="changeLog.version">
          <div class="mb-2 flex items-center gap-2 font-semibold">
            {{ changeLog.version }} -
            <span class="font-normal text-muted-foreground">
              {{ formatDate(changeLog.creationDate) }}
            </span>
          </div>

          <!-- eslint-disable vue/no-v-html -->
          <div class="richtext" v-html="changelogText(changeLog)" />
          <!-- eslint-enable vue/no-v-html -->
        </li>
      </ul>

      <Alert v-else variant="destructive" class="border-destructive/30 bg-destructive/10">
        <CircleX class="size-4" />
        <AlertDescription>{{ $t("extensionChangelog.noData") }}</AlertDescription>
      </Alert>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CircleX } from "lucide-vue-next";
import { formatDate } from "@/helpers/formatter";
import { useChangelogText } from "@/helpers/changelog";
import type { ExtensionWithChangelog } from "@/composables/useExtensionChangelogModal";

const changelogText = useChangelogText();

interface Props {
  show: boolean;
  extension: ExtensionWithChangelog | null;
}

const props = defineProps<Props>();
defineEmits<{
  close: [];
}>();

const changelogEntries = computed(() => {
  const cl = props.extension?.changelog;
  if (Array.isArray(cl)) {
    return cl;
  }
  return [];
});
</script>
