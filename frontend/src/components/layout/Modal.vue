<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && emit('close')">
    <DialogContent class="max-w-3xl">
      <DialogHeader v-if="!!$slots.icon || !!$slots.title">
        <div class="flex items-start gap-3">
          <div v-if="!!$slots.icon" class="mt-0.5 shrink-0">
            <slot name="icon" />
          </div>
          <DialogTitle v-if="!!$slots.title">
            <slot name="title" />
          </DialogTitle>
        </div>
      </DialogHeader>

      <slot name="content" />

      <DialogFooter v-if="!!$slots.footer">
        <slot name="footer" />
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

defineProps<{ show: boolean; closeXMark?: boolean }>();

const emit = defineEmits<{ close: [] }>();
</script>
