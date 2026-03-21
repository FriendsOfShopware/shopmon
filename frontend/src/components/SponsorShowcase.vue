<template>
  <section v-if="sponsors.length" class="sponsor-showcase" :class="{ compact }">
    <div v-if="title || description" class="sponsor-showcase-header">
      <component :is="titleTag" v-if="title" class="sponsor-showcase-title">
        {{ title }}
      </component>
      <p v-if="description" class="sponsor-showcase-description">
        {{ description }}
      </p>
    </div>

    <div class="sponsor-showcase-grid">
      <a
        v-for="sponsor in sponsors"
        :key="sponsor.name"
        :href="sponsor.url"
        class="sponsor-showcase-card"
        target="_blank"
        rel="noreferrer noopener"
      >
        <div class="sponsor-showcase-name">
          {{ sponsor.name }}
        </div>

        <p v-if="sponsor.description" class="sponsor-showcase-copy">
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

<style scoped>
.sponsor-showcase {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  width: 100%;
}

.sponsor-showcase-header {
  text-align: center;
}

.sponsor-showcase-title {
  color: var(--home-heading-color);
  font-size: 32px;
  margin-bottom: 1rem;
}

.sponsor-showcase-description {
  margin: 0 auto;
  max-width: 48rem;
}

.sponsor-showcase-grid {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.sponsor-showcase-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 2rem 1.5rem;
  color: inherit;
  text-align: center;
  text-decoration: none;
  background-color: var(--panel-background);
  border-radius: 0.375rem;
  box-shadow:
    0 1px 3px 0 rgba(0, 0, 0, 0.1),
    0 1px 2px 0 rgba(0, 0, 0, 0.06);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease;

  &:hover {
    transform: translateY(-2px);
  }
}

.sponsor-showcase-name {
  color: var(--primary-color);
  font-size: 1.125rem;
  font-weight: 600;
}

.sponsor-showcase-copy {
  margin: 0;
  color: var(--text-color-muted);
}

.compact {
  gap: 1.5rem;

  .sponsor-showcase-card {
    padding: 1.5rem 1rem;
  }
}

.dark .sponsor-showcase-card:hover {
  box-shadow: none;
}
</style>
