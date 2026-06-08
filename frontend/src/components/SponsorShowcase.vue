<template>
  <section v-if="sponsors.length" :class="['flex w-full flex-col', compact ? 'gap-4' : 'gap-8']">
    <div v-if="title || description" :class="compact ? '' : 'text-center'">
      <component
        :is="titleTag"
        v-if="title"
        :class="compact ? 'text-sm font-semibold' : 'text-2xl font-bold tracking-tight'"
      >
        {{ title }}
      </component>
      <p
        v-if="description"
        :class="[
          'text-muted-foreground',
          compact ? 'mt-1 text-xs' : 'mx-auto mt-2 max-w-2xl text-sm',
        ]"
      >
        {{ description }}
      </p>
    </div>

    <div :class="['grid gap-3', compact ? 'grid-cols-1' : 'grid-cols-1 sm:grid-cols-2']">
      <a
        v-for="sponsor in sponsors"
        :key="sponsor.name"
        :href="sponsor.url"
        class="group flex items-start gap-4 rounded-xl border bg-card p-4 transition-all duration-200 hover:border-primary/30 hover:shadow-sm"
        target="_blank"
        rel="noreferrer noopener"
      >
        <div class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
          <icon-fa6-solid:heart class="size-4 text-primary" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-1.5">
            <span class="font-semibold group-hover:text-primary transition-colors">{{
              sponsor.name
            }}</span>
            <icon-fa6-solid:arrow-up-right-from-square
              class="size-2.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
            />
          </div>
          <p v-if="sponsor.description" class="mt-1 text-sm leading-relaxed text-muted-foreground">
            {{ sponsor.description }}
          </p>
        </div>
      </a>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { Sponsor } from "@/data/sponsors";

withDefaults(
  defineProps<{
    sponsors: Sponsor[];
    title?: string;
    description?: string;
    titleTag?: string;
    compact?: boolean;
  }>(),
  {
    title: "",
    description: "",
    titleTag: "h2",
    compact: false,
  },
);
</script>
