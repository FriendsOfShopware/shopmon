<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="max-w-2xl">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <icon-fa6-solid:rotate />
          {{ $t("updateWizard.title") }}
        </DialogTitle>
      </DialogHeader>

      <div>
        <Select
          @update:model-value="(v) => v && $emit('versionSelected', String(v))"
        >
          <SelectTrigger class="mb-3 w-full">
            <SelectValue :placeholder="$t('updateWizard.selectVersion')" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="version in shopwareVersions" :key="version" :value="version">
              {{ version }}
            </SelectItem>
          </SelectContent>
        </Select>

        <template v-if="loading">
          <div class="py-4 text-center">
            {{ $t("common.loading") }}
            <icon-fa6-solid:rotate class="ml-1 inline animate-spin" />
          </div>
        </template>

        <div v-if="extensions" :class="{ 'opacity-20': loading }">
          <h2 class="mb-2 text-lg font-medium">
            {{ $t("updateWizard.extensionCompatibility") }}
          </h2>

          <ul class="list-none p-0">
            <li
              v-for="extension in extensions"
              :key="extension.name"
              class="flex gap-2 p-2 odd:bg-accent/50 hover:bg-accent"
            >
              <div class="mt-0.5 shrink-0">
                <icon-fa6-regular:circle v-if="!extension.active" class="size-4 text-muted-foreground" />
                <icon-fa6-solid:circle-info v-else-if="!extension.compatibility" class="size-4 text-warning" />
                <icon-fa6-solid:circle-xmark v-else-if="extension.compatibility.type == 'red'" class="size-4 text-destructive" />
                <icon-fa6-solid:rotate v-else-if="extension.compatibility.label === 'Available now'" class="size-4 text-info" />
                <icon-fa6-solid:circle-check v-else class="size-4 text-success" />
              </div>

              <div>
                <component
                  :is="extension.storeLink ? 'a' : 'span'"
                  v-bind="extension.storeLink ? { href: extension.storeLink, target: '_blank' } : {}"
                >
                  <strong>{{ extension.label }}</strong>
                </component>
                <span class="opacity-60"> ({{ extension.name }})</span>

                <div v-if="!extension.compatibility || !extension.storeLink" class="text-muted-foreground">
                  {{ $t("updateWizard.notInStore") }}
                </div>
                <div v-else>
                  {{ extension.compatibility.label }}
                </div>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

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

defineProps<Props>();
defineEmits<{
  close: [];
  versionSelected: [version: string];
}>();
</script>
