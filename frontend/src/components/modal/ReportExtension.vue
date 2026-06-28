<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <div class="flex items-start gap-3">
          <div class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-warning/10">
            <icon-fa6-solid:flag class="size-4 text-warning" />
          </div>
          <div>
            <DialogTitle>{{ $t("reportExtension.title") }}</DialogTitle>
            <DialogDescription class="mt-1">
              {{ $t("reportExtension.subtitle", { name: extension?.label ?? extension?.name }) }}
            </DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <div class="space-y-4">
        <div class="space-y-1.5">
          <Label>{{ $t("reportExtension.category") }}</Label>
          <Select v-model="category">
            <SelectTrigger class="w-full">
              <SelectValue :placeholder="$t('reportExtension.selectCategory')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="c in categories" :key="c" :value="c">
                {{ $t(`reportExtension.categories.${c}`) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div class="space-y-1.5">
          <Label>{{ $t("reportExtension.comment") }}</Label>
          <Textarea
            v-model="comment"
            rows="4"
            maxlength="2000"
            :placeholder="$t('reportExtension.commentPlaceholder')"
          />
        </div>

        <Alert class="border-muted bg-muted/40">
          <icon-fa6-solid:circle-info class="size-4" />
          <AlertDescription>{{ $t("reportExtension.reviewNote") }}</AlertDescription>
        </Alert>
      </div>

      <DialogFooter>
        <Button variant="ghost" @click="$emit('close')">
          {{ $t("common.cancel") }}
        </Button>
        <Button :disabled="!canSubmit || submitting" @click="submit">
          <icon-fa6-solid:flag class="mr-1.5 size-3" />
          {{ $t("reportExtension.submit") }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { api } from "@/helpers/api";
import { useAlert } from "@/composables/useAlert";

type ReportCategory = "performance" | "security" | "compatibility" | "stability" | "other";

interface ReportableExtension {
  name: string;
  label?: string;
}

const props = defineProps<{
  show: boolean;
  extension: ReportableExtension | null;
}>();

const emit = defineEmits<{
  close: [];
  submitted: [];
}>();

const { t } = useI18n();
const { success, error } = useAlert();

const categories: ReportCategory[] = [
  "performance",
  "security",
  "compatibility",
  "stability",
  "other",
];

const category = ref<ReportCategory | "">("");
const comment = ref("");
const submitting = ref(false);

const canSubmit = computed(() => category.value !== "" && comment.value.trim().length > 0);

watch(
  () => props.show,
  (open) => {
    if (!open) {
      category.value = "";
      comment.value = "";
      submitting.value = false;
    }
  },
);

async function submit() {
  if (!props.extension || !canSubmit.value) return;
  submitting.value = true;
  try {
    const { error: apiError } = await api.POST("/account/extensions/{extensionName}/report", {
      params: { path: { extensionName: props.extension.name } },
      body: { category: category.value as ReportCategory, comment: comment.value.trim() },
    });
    if (apiError) {
      throw new Error((apiError as { message?: string })?.message ?? "request failed");
    }
    success(t("reportExtension.success"));
    emit("submitted");
    emit("close");
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  } finally {
    submitting.value = false;
  }
}
</script>
