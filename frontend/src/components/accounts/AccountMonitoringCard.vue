<script setup lang="ts">
import { computed } from "vue";

import { Button } from "@/components/ui/button";
import CardSection from "@/components/CardSection.vue";
import IconUserShield from "~icons/fa6-solid/user-shield";

const props = defineProps<{
  /** One of ok | pending | unsupported | error | off. */
  state: string;
  openFindings: number;
  to: Record<string, unknown>;
}>();

// Zero open findings is not automatically good news: an unsupported endpoint, a
// failed read or a baseline that has not been taken yet all produce zero too,
// and rendering those green would claim an all-clear nobody checked.
const allClear = computed(() => props.state === "ok" && props.openFindings === 0);
const hasFindings = computed(() => props.state === "ok" && props.openFindings > 0);
</script>

<template>
  <CardSection :icon="IconUserShield" :title="$t('accounts.title')">
    <template #action>
      <Button as-child variant="ghost" size="sm" class="h-7 text-xs">
        <RouterLink :to="to">
          {{ $t("accounts.viewAll") }}
          <icon-fa6-solid:arrow-right class="ml-1 size-2.5" />
        </RouterLink>
      </Button>
    </template>

    <div
      v-if="state === 'off'"
      class="flex items-center gap-2 rounded-lg bg-muted px-3 py-2.5 text-sm text-muted-foreground"
    >
      <icon-fa6-solid:circle-info class="size-4 shrink-0" />
      {{ $t("accounts.disabled") }}
    </div>

    <div
      v-else-if="allClear"
      class="flex items-center gap-2 rounded-lg bg-success/10 px-3 py-2.5 text-sm text-success"
    >
      <icon-fa6-solid:circle-check class="size-4 shrink-0" />
      {{ $t("accounts.noFindings") }}
    </div>

    <RouterLink
      v-else-if="hasFindings"
      :to="to"
      class="flex items-center gap-2.5 rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm transition-colors hover:bg-destructive/10"
    >
      <icon-fa6-solid:user-shield class="size-4 shrink-0 text-destructive" />
      <span class="min-w-0 flex-1">
        {{ $t("accounts.pendingReview", { count: openFindings }) }}
      </span>
      <icon-fa6-solid:chevron-right class="size-2.5 shrink-0 text-muted-foreground" />
    </RouterLink>

    <RouterLink
      v-else
      :to="to"
      class="flex items-center gap-2.5 rounded-lg border border-warning/20 bg-warning/5 px-3 py-2.5 text-sm transition-colors hover:bg-warning/10"
    >
      <icon-fa6-solid:circle-exclamation class="size-4 shrink-0 text-warning" />
      <span class="min-w-0 flex-1">{{ $t("accounts.state." + state) }}</span>
      <icon-fa6-solid:chevron-right class="size-2.5 shrink-0 text-muted-foreground" />
    </RouterLink>
  </CardSection>
</template>
