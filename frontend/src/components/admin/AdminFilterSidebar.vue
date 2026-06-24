<template>
  <div class="flex h-full flex-col">
    <!-- Search -->
    <div v-if="searchable" class="px-4 pb-4">
      <Label class="mb-1.5 block text-xs font-medium text-muted-foreground">
        {{ $t("admin.filterSearch") }}
      </Label>
      <div class="relative">
        <icon-fa6-solid:magnifying-glass
          class="pointer-events-none absolute left-2.5 top-1/2 size-3 -translate-y-1/2 text-muted-foreground"
        />
        <Input
          :model-value="search"
          :placeholder="searchPlaceholder"
          class="h-9 w-full pl-8 text-sm"
          @update:model-value="$emit('update:search', String($event))"
        />
      </div>
    </div>

    <Separator v-if="searchable && groups.length" />

    <!-- Filter / sort groups -->
    <div class="flex-1 space-y-5 overflow-y-auto px-4 py-4">
      <div v-for="group in groups" :key="group.key">
        <Label class="mb-2 block text-xs font-medium text-muted-foreground">
          {{ group.label }}
        </Label>
        <div class="flex flex-col gap-1">
          <button
            v-for="opt in group.options"
            :key="String(opt.value)"
            type="button"
            :class="[
              'flex items-center justify-between rounded-md px-2.5 py-1.5 text-left text-sm transition-colors',
              modelValue[group.key] === opt.value
                ? 'bg-accent font-medium text-accent-foreground'
                : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
            ]"
            @click="select(group.key, opt.value)"
          >
            <span class="truncate">{{ opt.label }}</span>
            <icon-fa6-solid:check
              v-if="modelValue[group.key] === opt.value"
              class="size-3 shrink-0 text-primary"
            />
          </button>
        </div>
      </div>
    </div>

    <!-- Clear all -->
    <template v-if="activeCount > 0">
      <Separator />
      <div class="px-4 py-3">
        <Button
          variant="ghost"
          size="sm"
          class="w-full justify-start text-muted-foreground"
          @click="clearAll"
        >
          <icon-fa6-solid:xmark class="mr-1.5 size-3" />
          {{ $t("admin.clearFilters") }}
        </Button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { computed } from "vue";

export interface FilterOption {
  label: string;
  value: string;
}

export interface FilterGroup {
  /** Key into the modelValue record; identifies this filter dimension. */
  key: string;
  label: string;
  options: FilterOption[];
  /** Value that represents "no filter" for this group (used by active-count + clear). */
  defaultValue: string;
}

const props = defineProps<{
  /** Current selected value per group key. */
  modelValue: Record<string, string>;
  groups: FilterGroup[];
  searchable?: boolean;
  search?: string;
  searchPlaceholder?: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: Record<string, string>];
  "update:search": [value: string];
  change: [];
}>();

const activeCount = computed(() => {
  let count = 0;
  for (const g of props.groups) {
    if (props.modelValue[g.key] !== g.defaultValue) count++;
  }
  if (props.searchable && props.search) count++;
  return count;
});

function select(key: string, value: string) {
  if (props.modelValue[key] === value) return;
  emit("update:modelValue", { ...props.modelValue, [key]: value });
  emit("change");
}

function clearAll() {
  const reset: Record<string, string> = { ...props.modelValue };
  for (const g of props.groups) {
    reset[g.key] = g.defaultValue;
  }
  emit("update:modelValue", reset);
  if (props.searchable) emit("update:search", "");
  emit("change");
}

defineExpose({ activeCount });
</script>
