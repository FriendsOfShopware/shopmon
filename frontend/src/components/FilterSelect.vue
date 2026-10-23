<script setup lang="ts" generic="T extends string">
import { computed, ref, useId } from "vue";
import { useI18n } from "vue-i18n";
import { Check, ChevronsUpDown } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const props = defineProps<{
  id: string;
  modelValue: T;
  label: string;
  options: { value: T; label: string }[];
  searchable?: boolean;
  searchPlaceholder?: string;
  emptyMessage?: string;
}>();
const emit = defineEmits<{ "update:modelValue": [value: T] }>();
const { t } = useI18n();
const searchLabel = computed(() => props.searchPlaceholder ?? t("common.searchOptions"));
const open = ref(false);
const listId = useId();
const selectedLabel = computed(
  () =>
    props.options.find((option) => option.value === props.modelValue)?.label ?? props.modelValue,
);

function select(value: unknown) {
  const option = props.options.find((option) => option.value === value);
  if (!option) return;
  emit("update:modelValue", option.value);
  open.value = false;
}
</script>

<template>
  <Select v-if="!searchable" :model-value="modelValue" @update:model-value="select">
    <SelectTrigger :id="id" :aria-label="label">
      <SelectValue>{{ selectedLabel }}</SelectValue>
    </SelectTrigger>
    <SelectContent>
      <SelectItem v-for="option in options" :key="option.value" :value="option.value">
        {{ option.label }}
      </SelectItem>
    </SelectContent>
  </Select>
  <Popover v-else v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        :id="id"
        type="button"
        variant="outline"
        class="h-9 w-full min-w-0 justify-between px-3 font-normal"
        :aria-label="`${label}: ${selectedLabel}`"
        @keydown.down.prevent="open = true"
      >
        <span class="truncate">{{ selectedLabel }}</span>
        <ChevronsUpDown class="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
      </Button>
    </PopoverTrigger>
    <PopoverContent align="start" class="w-(--reka-popover-trigger-width) p-0" :aria-label="label">
      <Command :model-value="modelValue">
        <CommandInput
          :placeholder="searchLabel"
          :aria-label="searchLabel"
          role="combobox"
          aria-autocomplete="list"
          :aria-expanded="open"
          :aria-controls="listId"
          @keydown.enter.prevent
        />
        <CommandList :id="listId" :aria-label="label" class="max-h-64 p-1">
          <CommandEmpty>{{ emptyMessage ?? t("common.noMatchingOptions") }}</CommandEmpty>
          <CommandGroup>
            <CommandItem
              v-for="option in options"
              :key="option.value"
              :value="option.value"
              @select="select(option.value)"
            >
              <span class="min-w-0 flex-1 break-words">{{ option.label }}</span>
              <Check
                v-if="modelValue === option.value"
                class="size-4 shrink-0"
                aria-hidden="true"
              />
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </PopoverContent>
  </Popover>
</template>
