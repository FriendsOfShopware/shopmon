<template>
  <Dialog :open="open" @update:open="(v: boolean) => !v && emit('update:open', false)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("shop.createApiKeyTitle") }}</DialogTitle>
      </DialogHeader>
      <form id="apiKeyForm" class="space-y-4" @submit="onSubmit">
        <FormField v-slot="{ componentField }" name="apiKeyName">
          <FormItem>
            <FormLabel>{{ $t("common.name") }}</FormLabel>
            <FormControl>
              <Input v-bind="componentField" :placeholder="$t('shop.apiKeyPlaceholder')" />
            </FormControl>
            <FormMessage />
            <p class="text-xs text-muted-foreground">{{ $t("packages.apiKeyHelp") }}</p>
          </FormItem>
        </FormField>

        <div>
          <label class="text-sm font-medium">{{ $t("shop.scopes") }}</label>
          <p class="mb-2 text-xs text-muted-foreground">{{ $t("shop.scopesHelp") }}</p>
          <div class="mt-2 flex flex-col gap-2">
            <label
              v-for="scope in availableScopes"
              :key="scope.value"
              class="flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors hover:bg-muted"
            >
              <input
                type="checkbox"
                :value="scope.value"
                class="mt-1"
                :checked="selectedScopes.includes(scope.value)"
                @change="toggleScope(scope.value)"
              />
              <div>
                <span class="text-sm font-medium">{{ scope.label }}</span>
                <p class="text-xs text-muted-foreground">{{ scope.description }}</p>
              </div>
            </label>
          </div>
          <p v-if="scopeError" class="mt-1 text-sm text-destructive">{{ scopeError }}</p>
        </div>
      </form>
      <DialogFooter>
        <Button variant="outline" @click="emit('update:open', false)">{{
          $t("common.cancel")
        }}</Button>
        <Button type="submit" form="apiKeyForm" :disabled="isCreating">
          <icon-fa6-solid:key v-if="!isCreating" class="mr-1.5 size-3.5" />
          <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
          {{ $t("shop.createApiKey") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
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

type AvailableScope = components["schemas"]["ApiKeyScope"];

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
    const { data: result, error } = await api.POST(
      "/organizations/{orgId}/shops/{shopId}/api-keys",
      {
        params: { path: { orgId: props.orgId, shopId: props.shopId } },
        body: { name: values.apiKeyName, scopes: selectedScopes.value },
      },
    );
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
