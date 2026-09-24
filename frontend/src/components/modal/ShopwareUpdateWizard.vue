<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="flex max-h-[calc(100dvh-4rem)] max-w-2xl flex-col">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <icon-fa6-solid:rotate />
          {{ $t("updateWizard.title") }}
        </DialogTitle>
      </DialogHeader>

      <Select @update:model-value="(v) => v && $emit('versionSelected', String(v))">
        <SelectTrigger class="w-full shrink-0">
          <SelectValue :placeholder="$t('updateWizard.selectVersion')" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="version in shopwareVersions" :key="version" :value="version">
            {{ version }}
          </SelectItem>
        </SelectContent>
      </Select>

      <div
        v-if="extensions"
        class="flex shrink-0 flex-wrap items-center gap-x-4 gap-y-1"
        :class="{ 'opacity-20': loading }"
      >
        <h2 class="text-lg font-medium">
          {{ $t("updateWizard.extensionCompatibility") }}
        </h2>

        <div class="flex flex-wrap gap-x-3 gap-y-1 text-sm">
          <span
            v-for="entry in statusSummary"
            :key="entry.state"
            class="inline-flex items-center gap-1"
          >
            <component :is="entry.icon" class="size-3.5" :class="entry.iconClass" />
            {{ entry.count }} {{ $t(`updateWizard.summary.${entry.state}`) }}
          </span>
        </div>
      </div>

      <div class="-mx-6 min-h-0 overflow-y-auto px-6">
        <template v-if="loading">
          <div class="py-4 text-center">
            {{ $t("common.loading") }}
            <icon-fa6-solid:rotate class="ml-1 inline animate-spin" />
          </div>
        </template>

        <ul v-if="extensions" class="list-none p-0" :class="{ 'opacity-20': loading }">
          <li
            v-for="extension in extensions"
            :key="extension.name"
            class="flex gap-2 p-2 odd:bg-accent/50 hover:bg-accent"
          >
            <div class="mt-0.5 shrink-0">
              <component
                :is="statusMeta[getCompatibilityState(extension)].icon"
                class="size-4"
                :class="statusMeta[getCompatibilityState(extension)].iconClass"
              />
            </div>

            <div>
              <component
                :is="extension.storeLink ? 'a' : 'span'"
                v-bind="extension.storeLink ? { href: extension.storeLink, target: '_blank' } : {}"
              >
                <strong>{{ extension.label }}</strong>
              </component>
              <span class="opacity-60"> ({{ extension.name }})</span>

              <div
                v-if="!extension.compatibility || !extension.storeLink"
                class="text-muted-foreground"
              >
                {{ $t("updateWizard.notInStore") }}
              </div>
              <div v-else>
                {{ extension.compatibility.label }}
              </div>
            </div>
          </li>
        </ul>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { type Component, computed } from "vue";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { type CompatibilityState, getCompatibilityState } from "@/helpers/extensionCompatibility";
import IconCircle from "~icons/fa6-regular/circle";
import IconCircleCheck from "~icons/fa6-solid/circle-check";
import IconCircleInfo from "~icons/fa6-solid/circle-info";
import IconCircleXmark from "~icons/fa6-solid/circle-xmark";
import IconRotate from "~icons/fa6-solid/rotate";

interface ExtensionCompatibility {
  type: string;
  label: string;
}

interface Extension {
  name: string;
  label: string;
  active: boolean;
  compatibility?: ExtensionCompatibility;
  storeLink?: string | null;
}

interface Props {
  show: boolean;
  shopwareVersions: string[] | null;
  loading: boolean;
  extensions: Extension[] | null;
}

const props = defineProps<Props>();
defineEmits<{
  close: [];
  versionSelected: [version: string];
}>();

const statusMeta: Record<CompatibilityState, { icon: Component; iconClass: string }> = {
  incompatible: { icon: IconCircleXmark, iconClass: "text-destructive" },
  unknown: { icon: IconCircleInfo, iconClass: "text-warning" },
  updateAvailable: { icon: IconRotate, iconClass: "text-info" },
  compatible: { icon: IconCircleCheck, iconClass: "text-success" },
  inactive: { icon: IconCircle, iconClass: "text-muted-foreground" },
};

const statusSummary = computed(() => {
  if (!props.extensions) return [];

  const counts = new Map<CompatibilityState, number>();
  for (const extension of props.extensions) {
    const state = getCompatibilityState(extension);
    counts.set(state, (counts.get(state) ?? 0) + 1);
  }

  return (Object.keys(statusMeta) as CompatibilityState[])
    .filter((state) => counts.has(state))
    .map((state) => ({ state, count: counts.get(state) ?? 0, ...statusMeta[state] }));
});
</script>
