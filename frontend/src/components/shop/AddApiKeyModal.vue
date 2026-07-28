<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[485px]">
      <DialogHeader>
        <DialogTitle>{{ $t("shop.createApiKey") }}</DialogTitle>
      </DialogHeader>

      <form @submit="onSubmit" class="space-y-4">
        <FormField v-slot="{ componentField }" name="apiKeyName">
          <FormItem>
            <FormLabel>{{ $t("common.name") }}</FormLabel>
            <FormControl>
              <Input
                v-bind="componentField"
                :placeholder="$t('shop.apiKeyNamePlaceholder')"
                autocomplete="off"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <div>
          <FormLabel class="mb-2 block">{{ $t("shop.scopes") }}</FormLabel>
          <div class="space-y-2 rounded-md border p-3">
            <div
              v-for="scope in availableScopes"
              :key="scope.value"
              class="flex items-start space-x-2"
            >
              <input
                type="checkbox"
                :id="`scope-${scope.value}`"
                :value="scope.value"
                :checked="selectedScopes.includes(scope.value)"
                @change="toggleScope(scope.value)"
                class="mt-1 size-4 rounded border-gray-300 text-primary focus:ring-primary"
              />
              <label
                :for="`scope-${scope.value}`"
                class="cursor-pointer text-sm leading-tight select-none"
              >
                <div class="font-medium text-foreground">{{ scope.label }}</div>
                <div class="text-xs text-muted-foreground">{{ scope.description }}</div>
              </label>
            </div>
          </div>
          <p v-if="scopeError" class="mt-1 text-xs font-medium text-destructive">
            {{ scopeError }}
          </p>
        </div>
      </form>

      <DialogFooter>
        <Button variant="outline" @click="$emit('update:open', false)">
          {{ $t("common.cancel") }}
        </Button>
        <Button :disabled="isCreating" @click="onSubmit">
          <icon-fa6-solid:plus v-if="!isCreating" class="mr-1.5 size-3.5" />
          <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
          {{ $t("shop.createApiKey") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { createApiKey, type ApiKeyScope } from "@/api/generated";
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

type AvailableScope = ApiKeyScope;

const props = defineProps<{
  open: boolean;
  orgId: string;
  shopId: number;
  availableScopes: AvailableScope[];
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
  created: [token: string];
}>();

const { t } = useI18n();
const alert = useAlert();

const isCreating = ref(false);
const selectedScopes = ref<string[]>([]);
const scopeError = ref("");

const validationSchema = toTypedSchema(
  z.object({
    apiKeyName: z
      .string()
      .min(1, t("validation.nameRequired"))
      .max(100, t("validation.nameMaxLength")),
  }),
);

const { handleSubmit, resetForm } = useForm({
  validationSchema,
  initialValues: { apiKeyName: "" },
});

// Reset state every time the modal opens.
watch(
  () => props.open,
  (open) => {
    if (open) {
      selectedScopes.value = [];
      scopeError.value = "";
      resetForm({ values: { apiKeyName: "" } });
    }
  },
);

function toggleScope(value: string) {
  const idx = selectedScopes.value.indexOf(value);
  if (idx >= 0) {
    selectedScopes.value.splice(idx, 1);
  } else {
    selectedScopes.value.push(value);
  }
  scopeError.value = "";
}

const onSubmit = handleSubmit(async (values) => {
  if (selectedScopes.value.length === 0) {
    scopeError.value = t("validation.required", { field: t("shop.scopes") });
    return;
  }
  isCreating.value = true;
  try {
    const { data: result, error } = await createApiKey({
      path: { orgId: props.orgId, shopId: props.shopId },
      body: { name: values.apiKeyName, scopes: selectedScopes.value },
    });
    if (error) {
      alert.error(
        `${t("shop.failedCreateApiKey")}: ${(error as { message?: string }).message ?? ""}`,
      );
      return;
    }
    emit("update:open", false);
    emit("created", result?.token ?? "");
    alert.success(t("shop.apiKeyCreated"));
  } catch (error) {
    alert.error(
      `${t("shop.failedCreateApiKey")}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isCreating.value = false;
  }
});
</script>
