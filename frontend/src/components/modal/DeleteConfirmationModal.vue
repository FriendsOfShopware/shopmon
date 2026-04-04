<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <div class="flex items-start gap-3">
          <icon-fa6-solid:triangle-exclamation class="mt-0.5 size-5 shrink-0 text-destructive" aria-hidden="true" />
          <DialogTitle>{{ title }}</DialogTitle>
        </div>
      </DialogHeader>

      <div v-if="customMessage">
        {{ customMessage }}
      </div>

      <div v-else>
        <p v-html="$t('deleteModal.confirmText', { name: entityName })" />
        <p v-html="$t('deleteModal.consequence')" />
        <p v-if="customConsequence" class="text-muted-foreground">
          {{ customConsequence }}
        </p>
      </div>

      <div v-if="requirePassword" class="mt-4">
        <label for="deletePassword" class="mb-1 block font-medium">{{ $t("deleteModal.currentPassword") }}</label>
        <Input
          id="deletePassword"
          v-model="passwordValue"
          type="password"
          autocomplete="off"
          @input="$emit('password-change', passwordValue)"
        />
      </div>

      <DialogFooter>
        <Button variant="ghost" @click="$emit('close')">
          {{ $t("common.cancel") }}
        </Button>
        <Button
          variant="destructive"
          :disabled="isLoading || (requirePassword && !passwordValue)"
          @click="$emit('confirm')"
        >
          <icon-fa6-solid:trash v-if="!isLoading" class="mr-1 size-4" />
          <icon-line-md:loading-twotone-loop v-if="isLoading" class="mr-1 size-4" />
          {{ confirmButtonText || $t("common.delete") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface Props {
  show: boolean;
  title: string;
  entityName: string;
  customMessage?: string;
  customConsequence?: string;
  requirePassword?: boolean;
  isLoading?: boolean;
  confirmButtonText?: string;
  reversedButtons?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  customMessage: "",
  customConsequence: "",
  requirePassword: false,
  isLoading: false,
  confirmButtonText: "",
  reversedButtons: false,
});

defineEmits<{
  close: [];
  confirm: [];
  "password-change": [password: string];
}>();

const passwordValue = ref("");

watch(
  () => props.show,
  (newShow) => {
    if (!newShow) {
      passwordValue.value = "";
    }
  },
);
</script>
