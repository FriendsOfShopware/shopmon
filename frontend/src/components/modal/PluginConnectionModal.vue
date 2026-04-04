<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ $t("pluginModal.title") }}</DialogTitle>
      </DialogHeader>

      <div>
        <p class="mb-3 text-muted-foreground">{{ $t("pluginModal.description") }}</p>
        <Textarea
          v-model="pluginBase64Value"
          rows="4"
          placeholder="eyJ1cmwiOiJodHRwczpcL1wvZGVtby5mb3MuZ2ciLCJjbGllbnRJZCI6..."
          @input="$emit('update:base64', pluginBase64Value)"
        />
        <p v-if="error" class="mt-1 text-sm text-destructive">
          {{ error }}
        </p>
      </div>

      <DialogFooter>
        <Button variant="ghost" @click="$emit('close')">
          {{ $t("common.cancel") }}
        </Button>
        <Button @click="$emit('import')">
          {{ $t("pluginModal.importData") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

interface Props {
  show: boolean;
  base64: string;
  error?: string;
}

const props = withDefaults(defineProps<Props>(), {
  error: "",
});

defineEmits<{
  close: [];
  import: [];
  "update:base64": [value: string];
}>();

const pluginBase64Value = ref(props.base64);

watch(
  () => props.base64,
  (newValue) => {
    pluginBase64Value.value = newValue;
  },
);

watch(
  () => props.show,
  (newShow) => {
    if (!newShow) {
      pluginBase64Value.value = "";
    }
  },
);
</script>
