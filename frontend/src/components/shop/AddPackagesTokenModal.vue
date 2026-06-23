<template>
  <Dialog :open="open" @update:open="(v: boolean) => !v && emit('update:open', false)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("packages.addTokenTitle") }}</DialogTitle>
      </DialogHeader>
      <form id="packagesTokenForm" class="space-y-4" @submit="onSubmit">
        <FormField v-slot="{ componentField }" name="packagesToken">
          <FormItem>
            <FormLabel>{{ $t("packages.tokenLabel") }}</FormLabel>
            <FormControl>
              <Input v-bind="componentField" :placeholder="$t('packages.tokenPlaceholder')" />
            </FormControl>
            <FormMessage />
            <p class="text-xs text-muted-foreground">{{ $t("packages.tokenHelp") }}</p>
          </FormItem>
        </FormField>
      </form>
      <DialogFooter>
        <Button variant="outline" @click="emit('update:open', false)">{{
          $t("common.cancel")
        }}</Button>
        <Button type="submit" form="packagesTokenForm" :disabled="isCreating">
          <icon-fa6-solid:plus v-if="!isCreating" class="mr-1.5 size-3.5" />
          <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
          {{ $t("packages.addToken") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const props = defineProps<{
  open: boolean;
  orgId: string;
  shopId: number;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
  created: [];
}>();

const { t } = useI18n();
const alert = useAlert();

const isCreating = ref(false);

const validationSchema = toTypedSchema(
  z.object({
    packagesToken: z.string().min(1, t("validation.tokenRequired")),
  }),
);

const { handleSubmit, resetForm } = useForm({
  validationSchema,
  initialValues: { packagesToken: "" },
});

// Reset the form every time the modal opens.
watch(
  () => props.open,
  (open) => {
    if (open) resetForm({ values: { packagesToken: "" } });
  },
);

const onSubmit = handleSubmit(async (values) => {
  isCreating.value = true;
  try {
    const { error } = await api.POST("/organizations/{orgId}/shops/{shopId}/packages-tokens", {
      params: { path: { orgId: props.orgId, shopId: props.shopId } },
      body: { token: values.packagesToken },
    });
    if (error) {
      alert.error(`${t("packages.failedAdd")}: ${(error as { message?: string }).message ?? ""}`);
      return;
    }
    emit("update:open", false);
    emit("created");
    alert.success(t("packages.tokenAdded"));
  } catch (error) {
    alert.error(`${t("packages.failedAdd")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isCreating.value = false;
  }
});
</script>
