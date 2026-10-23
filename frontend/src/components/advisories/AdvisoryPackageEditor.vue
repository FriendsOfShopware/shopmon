<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { AdvisoryPackageInput } from "@/api/generated";
import { useId } from "vue";

const packages = defineModel<AdvisoryPackageInput[]>({ required: true });
const id = useId();

function addPackage() {
  packages.value = [...packages.value, { packageName: "", affectedVersions: "" }];
}

function removePackage(index: number) {
  packages.value = packages.value.filter((_, i) => i !== index);
}
</script>

<template>
  <div class="space-y-3">
    <div
      v-for="(pkg, index) in packages"
      :key="index"
      class="flex flex-wrap items-end gap-2 rounded-md bg-muted px-2.5 py-2.5"
    >
      <div class="min-w-44 flex-1 space-y-1.5">
        <Label :for="`${id}-${index}-name`" class="text-xs">{{ $t("admin.packageName") }}</Label>
        <Input
          :id="`${id}-${index}-name`"
          v-model="pkg.packageName"
          class="bg-background font-mono text-xs"
          placeholder="shopware/core"
        />
      </div>
      <div class="min-w-44 flex-1 space-y-1.5">
        <Label :for="`${id}-${index}-versions`" class="text-xs">{{
          $t("advisories.affectedVersions")
        }}</Label>
        <Input
          :id="`${id}-${index}-versions`"
          v-model="pkg.affectedVersions"
          class="bg-background font-mono text-xs"
          placeholder=">=6.7.0.0,<6.7.10.1"
        />
      </div>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        :disabled="packages.length <= 1"
        :aria-label="$t('admin.removePackage')"
        :title="$t('admin.removePackage')"
        @click="removePackage(index)"
      >
        <icon-fa6-solid:trash class="size-3.5" />
      </Button>
    </div>
    <Button type="button" variant="outline" size="sm" @click="addPackage">
      <icon-fa6-solid:plus class="mr-1.5 size-3" />
      {{ $t("admin.addPackage") }}
    </Button>
  </div>
</template>
