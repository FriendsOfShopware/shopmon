<template>
  <Dialog :open="open" @update:open="(v: boolean) => !v && emit('update:open', false)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("organization.addMember") }}</DialogTitle>
      </DialogHeader>
      <form id="addMemberForm" class="space-y-4" @submit="onSubmit">
        <FormField v-slot="{ componentField }" name="email">
          <FormItem>
            <FormLabel>{{ $t("common.email") }}</FormLabel>
            <FormControl>
              <Input v-bind="componentField" autocomplete="email" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>
        <FormField v-slot="{ componentField }" name="role">
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
        <Button :disabled="submitting" type="submit" form="addMemberForm">
          <icon-fa6-solid:plus v-if="!submitting" class="mr-1.5 size-3.5" />
          <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
          {{ $t("common.add") }}
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
import { Input } from "@/components/ui/input";
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

const props = defineProps<{ open: boolean; submitting: boolean }>();
const emit = defineEmits<{
  "update:open": [value: boolean];
  submit: [values: { email: string; role: "member" | "admin" }];
}>();

const schema = toTypedSchema(
  z.object({
    email: z.string().email(t("validation.emailInvalid")).min(1, t("validation.emailRequired")),
    role: z.enum(["member", "admin"], { message: t("validation.roleInvalid") }),
  }),
);

const { handleSubmit, resetForm } = useForm({
  validationSchema: schema,
  initialValues: { email: "", role: "member" as const },
});

// Reset state every time the modal opens.
watch(
  () => props.open,
  (open) => {
    if (open) resetForm({ values: { email: "", role: "member" } });
  },
);

const onSubmit = handleSubmit((values) => {
  emit("submit", values);
});
</script>
