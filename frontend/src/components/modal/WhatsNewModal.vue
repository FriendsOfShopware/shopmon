<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="max-w-2xl">
      <DialogHeader>
        <div class="flex items-start gap-3">
          <div class="flex size-12 items-center justify-center rounded-full bg-gradient-to-br from-primary to-primary/80 text-white">
            <icon-fa6-solid:rocket class="size-5" />
          </div>
          <DialogTitle>{{ $t("whatsNew.title") }}</DialogTitle>
        </div>
      </DialogHeader>

      <div class="space-y-4">
        <div class="mb-4">
          <span class="inline-flex items-center rounded-full bg-primary/10 px-3 py-1 text-xs font-bold uppercase tracking-wide text-primary">
            {{ $t("whatsNew.badge") }}
          </span>
          <h4 class="mt-3 text-xl font-semibold">{{ $t("whatsNew.heroTitle") }}</h4>
          <p class="mt-2 text-muted-foreground">{{ $t("whatsNew.heroDesc") }}</p>
        </div>

        <ul class="grid gap-3 p-0">
          <li class="grid grid-cols-[1.25rem_1fr] gap-3 rounded-xl border bg-accent/50 p-4">
            <icon-fa6-solid:bolt class="mt-0.5 size-4 text-primary" />
            {{ $t("whatsNew.featureCdn") }}
          </li>
          <li class="grid grid-cols-[1.25rem_1fr] gap-3 rounded-xl border bg-accent/50 p-4">
            <icon-fa6-solid:arrows-rotate class="mt-0.5 size-4 text-primary" />
            {{ $t("whatsNew.featureSync") }}
          </li>
          <li class="grid grid-cols-[1.25rem_1fr] gap-3 rounded-xl border bg-accent/50 p-4">
            <icon-fa6-solid:shield-halved class="mt-0.5 size-4 text-primary" />
            {{ $t("whatsNew.featureValidation") }}
          </li>
        </ul>

        <div class="rounded-xl border border-primary/20 bg-primary/5 p-4">
          <p class="mb-1 font-bold">{{ $t("whatsNew.quickSetup") }}</p>
          <p class="m-0 text-muted-foreground">{{ $t("whatsNew.quickSetupDesc") }}</p>
        </div>

        <div v-if="sponsors.length" class="mt-4">
          <p class="mb-1 font-bold">{{ $t("whatsNew.sponsors") }}</p>
          <p class="mb-4 text-muted-foreground">{{ $t("whatsNew.sponsorsDesc") }}</p>
          <SponsorShowcase :sponsors="sponsors" compact />
        </div>
      </div>

      <DialogFooter>
        <UiButton :to="{ name: 'account.shop.list' }" @click="$emit('close')">
          <icon-fa6-solid:folder-open class="mr-1 size-4" />
          {{ $t("whatsNew.openProjects") }}
        </UiButton>

        <UiButton :to="{ name: 'account.docs' }" @click="$emit('close')">
          <icon-fa6-solid:book class="mr-1 size-4" />
          {{ $t("nav.documentation") }}
        </UiButton>

        <UiButton type="button" variant="primary" @click="$emit('close')">
          {{ $t("whatsNew.close") }}
        </UiButton>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import SponsorShowcase from "@/components/SponsorShowcase.vue";
import { sponsors } from "@/data/sponsors";

defineProps<{ show: boolean }>();
defineEmits<{ close: [] }>();
</script>
