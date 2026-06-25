<template>
  <Dialog :open="open" @update:open="(v: boolean) => !v && emit('update:open', false)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("organization.changeMemberRole") }}</DialogTitle>
      </DialogHeader>
      <form id="changeRoleForm" class="space-y-4" @submit="onSubmit">
        <FormField v-slot="{ componentField }" name="changeRole">
          <FormItem>
            <FormLabel>{{ $t("common.role") }}</FormLabel>
            <Select v-bind="componentField">
              <FormControl>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                <SelectItem value="member">{{ $t("organization.roleMember") }}</SelectItem>
                <SelectItem value="admin">{{ $t("organization.roleAdmin") }}</SelectItem>
              </SelectContent>
            </Select>
            <FormMessage />
          </FormItem>
        </FormField>
      </form>
      <DialogFooter>
        <Button variant="outline" @click="emit('update:open', false)">{{
          $t("common.cancel")
        }}</Button>
        <Button :disabled="submitting" type="submit" form="changeRoleForm">
          <icon-fa6-solid:floppy-disk v-if="!submitting" class="mr-1.5 size-3.5" />
          <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
          {{ $t("common.save") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toTypedSchema } from "@vee-validate/zod";
import { useForm } from "vee-validate";
import { watch } from "vue";
import { useI18n } from "vue-i18n";
import { z } from "zod";

const { t } = useI18n();

const props = defineProps<{ open: boolean; submitting: boolean; role: "member" | "admin" }>();
const emit = defineEmits<{
  "update:open": [value: boolean];
  submit: [role: "member" | "admin"];
}>();

const schema = toTypedSchema(
  z.object({
    changeRole: z.enum(["member", "admin"], { message: t("validation.roleInvalid") }),
  }),
);

const { handleSubmit, setValues } = useForm({
  validationSchema: schema,
  initialValues: { changeRole: "member" as const },
});

watch(
  () => [props.open, props.role] as const,
  ([isOpen, role]) => {
    if (isOpen) setValues({ changeRole: role });
  },
  { immediate: true },
);

const onSubmit = handleSubmit((values) => {
  emit("submit", values.changeRole);
});
</script>
