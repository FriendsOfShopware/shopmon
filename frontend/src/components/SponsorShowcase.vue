<template>
  <section v-if="sponsors.length" :class="['flex w-full flex-col', compact ? 'gap-6' : 'gap-8']">
    <div v-if="title || description" class="text-center">
      <component :is="titleTag" v-if="title" class="mb-4 text-3xl font-bold">
        {{ title }}
      </component>
      <p v-if="description" class="mx-auto max-w-3xl text-muted-foreground">
        {{ description }}
      </p>
    </div>

    <div class="grid gap-6 grid-cols-[repeat(auto-fit,minmax(220px,1fr))]">
      <a
        v-for="sponsor in sponsors"
        :key="sponsor.name"
        :href="sponsor.url"
        :class="['flex flex-col items-center justify-center gap-4 rounded-md bg-card text-center shadow transition-transform hover:-translate-y-0.5', compact ? 'px-4 py-6' : 'px-6 py-8']"
        target="_blank"
        rel="noreferrer noopener"
      >
        <div class="text-lg font-semibold text-primary">
          {{ sponsor.name }}
        </div>

        <p v-if="sponsor.description" class="m-0 text-muted-foreground">
          {{ sponsor.description }}
        </p>
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
