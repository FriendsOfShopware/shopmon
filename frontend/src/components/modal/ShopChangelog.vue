<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="max-w-2xl">
      <DialogHeader>
        <DialogTitle>
          {{
            $t("changelogModal.title", {
              date: changelog?.date ? formatDateTime(changelog.date) : "",
            })
          }}
        </DialogTitle>
      </DialogHeader>

      <div>
        <template v-if="changelog?.oldShopwareVersion && changelog?.newShopwareVersion">
          <span
            v-html="
              $t('changelogModal.shopwareUpdate', {
                oldVersion: changelog.oldShopwareVersion,
                newVersion: changelog.newShopwareVersion,
              })
            "
          />
        </template>

        <div v-if="changelog?.extensions?.length">
          <h2 class="mb-2 mt-2 text-lg font-medium">{{ $t("changelogModal.pluginChangelog") }}</h2>

          <ul class="list-disc">
            <li v-for="extension in changelog?.extensions" :key="extension.name" class="ml-4 mb-2">
              <div class="font-medium">
                {{ extension.label }}
                <span class="font-normal text-muted-foreground">({{ extension.name }})</span>
              </div>

              {{ extension.state }}
              <template v-if="extension.state === 'installed' && extension.active">
                {{ $t("changelogModal.andActivated") }}
              </template>
              <template v-if="extension.state === 'updated'">
                {{
                  $t("changelogModal.fromTo", {
                    oldVersion: extension.oldVersion,
                    newVersion: extension.newVersion,
                  })
                }}
              </template>
              <template v-else>
                {{ extension.newVersion }}
                <template v-if="!extension.newVersion">
                  {{ extension.oldVersion }}
                </template>
              </template>

              <ul
                v-if="extension.state === 'updated' && extension.changelog?.length"
                class="mt-2 list-none space-y-4 p-0"
              >
                <li v-for="entry in extension.changelog" :key="entry.version">
                  <div class="mb-1 flex items-center gap-2 font-semibold">
                    {{ entry.version }} -
                    <span class="font-normal text-muted-foreground">{{
                      formatDateTime(entry.creationDate)
                    }}</span>
                  </div>
                  <!-- eslint-disable vue/no-v-html -->
                  <div class="prose prose-sm dark:prose-invert max-w-none" v-html="entry.text" />
                  <!-- eslint-enable vue/no-v-html -->
                </li>
              </ul>
            </li>
          </ul>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { formatDateTime } from "@/helpers/formatter";
import type { components } from "@/types/api";

type AccountChangelog = components["schemas"]["AccountChangelog"];

interface Props {
  show: boolean;
  changelog: AccountChangelog | null;
}

defineProps<Props>();
defineEmits<{
  close: [];
}>();
</script>
