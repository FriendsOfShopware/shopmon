<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="max-w-2xl">
      <DialogHeader>
        <DialogTitle>
          {{ $t("extensionChangelog.title") }} -
          <span class="font-normal text-muted-foreground">{{ extension?.name }}</span>
        </DialogTitle>
      </DialogHeader>

      <ul v-if="changelogEntries.length > 0" class="list-none space-y-6 p-0">
        <li v-for="changeLog in changelogEntries" :key="changeLog.version">
          <div class="mb-2 flex items-center gap-2 font-semibold">
            <span
              v-if="!changeLog.isCompatible"
              :title="$t('extensionChangelog.notCompatible')"
            >
              <icon-fa6-solid:circle-info class="size-4 text-warning" />
            </span>

            {{ changeLog.version }} -
            <span class="font-normal text-muted-foreground">
              {{ formatDate(changeLog.creationDate) }}
            </span>
          </div>

          <!-- eslint-disable vue/no-v-html -->
          <div class="prose prose-sm dark:prose-invert max-w-none" v-html="changeLog.text" />
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
import type { ExtensionWithChangelog } from "@/composables/useExtensionChangelogModal";

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
