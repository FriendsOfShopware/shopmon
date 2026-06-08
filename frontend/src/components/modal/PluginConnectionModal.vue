<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent>
      <DialogHeader>
        <div class="flex items-start gap-3">
          <div class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10">
            <icon-fa6-solid:plug class="size-4 text-primary" />
          </div>
          <div>
            <DialogTitle>{{ $t("pluginModal.title") }}</DialogTitle>
            <DialogDescription class="mt-1">
              {{ $t("pluginModal.subtitle") }}
            </DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <ol class="space-y-4 text-sm">
        <li class="flex items-start gap-3">
          <span
            class="flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-bold text-muted-foreground"
            >1</span
          >
          <div>
            <p class="font-medium">{{ $t("pluginModal.step1Title") }}</p>
            <p class="mt-0.5 text-xs text-muted-foreground">
              <i18n-t keypath="pluginModal.step1Desc" tag="span">
                <template #link>
                  <a
                    href="https://store.shopware.com/en/frosh99285362831f/shopmon.html"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-primary hover:underline"
                    >FroshShopmon</a
                  >
                </template>
              </i18n-t>
            </p>
          </div>
        </li>

        <li class="flex items-start gap-3">
          <span
            class="flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-bold text-muted-foreground"
            >2</span
          >
          <div>
            <p class="font-medium">{{ $t("pluginModal.step2Title") }}</p>
            <p class="mt-0.5 text-xs text-muted-foreground">{{ $t("pluginModal.step2Desc") }}</p>
          </div>
        </li>

        <li class="flex items-start gap-3">
          <span
            class="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground"
            >3</span
          >
          <p class="font-medium">{{ $t("pluginModal.step3Title") }}</p>
        </li>
      </ol>

      <Textarea
        v-model="pluginBase64Value"
        rows="3"
        class="font-mono text-xs"
        placeholder="eyJ1cmwiOiJodHRwczovL215LXNob3AuZXhhbXBsZS5jb20iLC..."
        @input="$emit('update:base64', pluginBase64Value)"
      />

      <Alert v-if="error" variant="destructive">
        <icon-fa6-solid:circle-exclamation class="size-4" />
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <DialogFooter>
        <Button variant="ghost" @click="$emit('close')">
          {{ $t("common.cancel") }}
        </Button>
        <Button @click="$emit('import')" :disabled="!pluginBase64Value.trim()">
          <icon-fa6-solid:file-import class="mr-1.5 size-3" />
          {{ $t("pluginModal.importData") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Alert, AlertDescription } from "@/components/ui/alert";

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
