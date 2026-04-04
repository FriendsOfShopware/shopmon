<template>
  <Dialog :open="show" @update:open="(v: boolean) => !v && $emit('close')">
    <DialogContent class="max-w-lg gap-0 p-0">
      <!-- Header with gradient -->
      <div class="rounded-t-lg bg-gradient-to-br from-primary to-primary/80 px-6 pb-6 pt-8 text-white">
        <div class="mb-3 flex items-center gap-2">
          <div class="flex size-8 items-center justify-center rounded-full bg-white/15">
            <icon-fa6-solid:rocket class="size-3.5" />
          </div>
          <Badge class="border-white/20 bg-white/15 text-xs text-white">
            {{ $t("whatsNew.badge") }}
          </Badge>
        </div>
        <h2 class="text-xl font-bold leading-tight">{{ $t("whatsNew.heroTitle") }}</h2>
        <p class="mt-2 text-sm leading-relaxed text-white/70">{{ $t("whatsNew.heroDesc") }}</p>
      </div>

      <!-- Content -->
      <div class="space-y-4 px-6 py-5">
        <!-- Feature list -->
        <div class="space-y-2">
          <div v-for="feature in features" :key="feature.text" class="flex items-start gap-3 rounded-lg border px-3 py-2.5">
            <div class="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-md bg-primary/10">
              <component :is="feature.icon" class="size-3 text-primary" />
            </div>
            <span class="text-sm">{{ feature.text }}</span>
          </div>
        </div>

        <!-- Quick setup -->
        <div class="rounded-lg bg-muted/50 p-4">
          <h4 class="mb-1 text-sm font-semibold">{{ $t("whatsNew.quickSetup") }}</h4>
          <p class="text-xs leading-relaxed text-muted-foreground">{{ $t("whatsNew.quickSetupDesc") }}</p>
        </div>

        <!-- Sponsors -->
        <div v-if="sponsors.length">
          <h4 class="mb-1 text-sm font-semibold">{{ $t("whatsNew.sponsors") }}</h4>
          <p class="mb-3 text-xs text-muted-foreground">{{ $t("whatsNew.sponsorsDesc") }}</p>
          <SponsorShowcase :sponsors="sponsors" compact />
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end border-t px-6 py-4">
        <Button size="sm" @click="$emit('close')">
          {{ $t("whatsNew.close") }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import SponsorShowcase from "@/components/SponsorShowcase.vue";
import { sponsors } from "@/data/sponsors";
import { useI18n } from "vue-i18n";

import FaBolt from "~icons/fa6-solid/bolt";
import FaArrowsRotate from "~icons/fa6-solid/arrows-rotate";
import FaShieldHalved from "~icons/fa6-solid/shield-halved";

const { t } = useI18n();

defineProps<{ show: boolean }>();
defineEmits<{ close: [] }>();

const features = [
  { icon: FaBolt, text: t("whatsNew.featureCdn") },
  { icon: FaArrowsRotate, text: t("whatsNew.featureSync") },
  { icon: FaShieldHalved, text: t("whatsNew.featureValidation") },
];
</script>
