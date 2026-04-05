<script setup lang="ts">
import { ref, useAttrs } from "vue";
import { Input } from "@/components/ui/input";

defineOptions({ inheritAttrs: false });

defineProps<{
  placeholder?: string;
  autocomplete?: string;
  id?: string;
}>();

const attrs = useAttrs();
const modelValue = defineModel<string>();
const visible = ref(false);
</script>

<template>
  <div class="relative">
    <Input
      v-model="modelValue"
      v-bind="attrs"
      :type="visible ? 'text' : 'password'"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      :id="id"
      class="pr-10"
    />
    <button
      type="button"
      class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground opacity-50 hover:opacity-100 transition-opacity"
      @click="visible = !visible"
    >
      <icon-fa6-solid:eye v-if="!visible" class="size-3.5" />
      <icon-fa6-solid:eye-slash v-else class="size-3.5" />
    </button>
  </div>
</template>
