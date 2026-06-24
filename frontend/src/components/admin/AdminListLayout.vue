<template>
  <div class="space-y-4">
    <!-- Header row: title + mobile filter trigger -->
    <div class="flex items-center justify-between gap-3">
      <h1 class="text-2xl font-bold tracking-tight">{{ title }}</h1>

      <!-- Mobile: open filters in a drawer -->
      <Sheet v-model:open="drawerOpen">
        <SheetTrigger as-child>
          <Button variant="outline" size="sm" class="lg:hidden">
            <icon-fa6-solid:sliders class="mr-1.5 size-3.5" />
            {{ $t("admin.filters") }}
            <Badge
              v-if="activeCount > 0"
              class="ml-1.5 size-4 justify-center rounded-full bg-primary p-0 text-[10px] text-primary-foreground"
            >
              {{ activeCount }}
            </Badge>
          </Button>
        </SheetTrigger>
        <SheetContent side="left" class="w-80 p-0">
          <SheetHeader class="border-b px-4 py-3 text-left">
            <SheetTitle class="text-base">{{ $t("admin.filters") }}</SheetTitle>
          </SheetHeader>
          <slot name="filters" :in-drawer="true" />
        </SheetContent>
      </Sheet>
    </div>

    <div class="flex flex-col gap-6 lg:flex-row lg:items-start">
      <!-- Desktop: sticky sidebar -->
      <aside class="hidden w-64 shrink-0 lg:block">
        <div class="sticky top-16 rounded-xl border bg-card py-4">
          <div class="px-4 pb-3">
            <span class="flex items-center gap-1.5 text-sm font-semibold">
              <icon-fa6-solid:sliders class="size-3.5 text-muted-foreground" />
              {{ $t("admin.filters") }}
            </span>
          </div>
          <Separator />
          <div class="pt-4">
            <slot name="filters" :in-drawer="false" />
          </div>
        </div>
      </aside>

      <!-- Content -->
      <div class="min-w-0 flex-1">
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { ref } from "vue";

defineProps<{
  title: string;
  activeCount: number;
}>();

const drawerOpen = ref(false);

defineExpose({ close: () => (drawerOpen.value = false) });
</script>
