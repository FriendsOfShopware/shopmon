<template>
  <nav :class="['breadcrumbs', `breadcrumbs-${size}`]" aria-label="breadcrumb">
    <div class="breadcrumbs-mobile">
      <template v-for="segment in mobileSegments" :key="segment.key">
        <span v-if="segment.kind === 'separator'" class="breadcrumbs-separator" aria-hidden="true">
          <svg width="24" height="24" fill="none" viewBox="0 0 24 24">
            <path
              stroke="currentColor"
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.5"
              d="M10.75 8.75L14.25 12L10.75 15.25"
            />
          </svg>
        </span>

        <span v-else-if="segment.kind === 'ellipsis'" class="breadcrumbs-ellipsis" aria-hidden="true">
          ...
        </span>

        <component
          :is="getSegmentTag(segment)"
          v-else
          v-bind="getSegmentBindings(segment)"
          :class="segment.isCurrent ? 'breadcrumbs-current' : 'breadcrumbs-link'"
          :aria-current="segment.isCurrent ? 'page' : undefined"
        >
          <component :is="segment.item.icon" v-if="segment.item.icon" class="breadcrumbs-icon" />
          <span class="breadcrumbs-label">{{ segment.item.label }}</span>
        </component>
      </template>
    </div>

    <div class="breadcrumbs-desktop">
      <template v-for="segment in desktopSegments" :key="segment.key">
        <span v-if="segment.kind === 'separator'" class="breadcrumbs-separator" aria-hidden="true">
          <svg width="24" height="24" fill="none" viewBox="0 0 24 24">
            <path
              stroke="currentColor"
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.5"
              d="M10.75 8.75L14.25 12L10.75 15.25"
            />
          </svg>
        </span>

        <component
          :is="getSegmentTag(segment)"
          v-else
          v-bind="getSegmentBindings(segment)"
          :class="segment.isCurrent ? 'breadcrumbs-current' : 'breadcrumbs-link'"
          :aria-current="segment.isCurrent ? 'page' : undefined"
        >
          <component :is="segment.item.icon" v-if="segment.item.icon" class="breadcrumbs-icon" />
          <span class="breadcrumbs-label">{{ segment.item.label }}</span>
        </component>
      </template>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { BreadcrumbItem } from "@/components/layout/breadcrumbs";

type BreadcrumbSize = "sm" | "base";

interface ItemSegment {
  kind: "item";
  key: string;
  item: BreadcrumbItem;
  isCurrent: boolean;
}

interface SeparatorSegment {
  kind: "separator";
  key: string;
}

interface EllipsisSegment {
  kind: "ellipsis";
  key: string;
}

type BreadcrumbSegment = ItemSegment | SeparatorSegment | EllipsisSegment;

const props = withDefaults(
  defineProps<{
    items: BreadcrumbItem[];
    size?: BreadcrumbSize;
  }>(),
  {
    size: "base",
  },
);

const normalizedItems = computed(() => {
  return props.items.filter((item) => item.label.trim().length > 0);
});

const desktopSegments = computed(() => {
  return buildSegments(normalizedItems.value);
});

const mobileSegments = computed(() => {
  if (normalizedItems.value.length <= 2) {
    return buildSegments(normalizedItems.value);
  }

  const parentItem = normalizedItems.value.at(-2);
  const currentItem = normalizedItems.value.at(-1);

  if (!parentItem || !currentItem) {
    return buildSegments(normalizedItems.value);
  }

  return [
    { kind: "ellipsis", key: "mobile-ellipsis" },
    { kind: "separator", key: "mobile-separator-leading" },
    { kind: "item", key: "mobile-parent", item: parentItem, isCurrent: false },
    { kind: "separator", key: "mobile-separator-trailing" },
    { kind: "item", key: "mobile-current", item: currentItem, isCurrent: true },
  ] satisfies BreadcrumbSegment[];
});

function buildSegments(items: BreadcrumbItem[]): BreadcrumbSegment[] {
  return items.flatMap((item, index) => {
    const isCurrent = index === items.length - 1;
    const segments: BreadcrumbSegment[] = [
      {
        kind: "item",
        key: `item-${index}-${item.label}`,
        item,
        isCurrent,
      },
    ];

    if (!isCurrent) {
      segments.push({
        kind: "separator",
        key: `separator-${index}`,
      });
    }

    return segments;
  });
}

function getSegmentTag(segment: ItemSegment) {
  if (segment.isCurrent || (!segment.item.to && !segment.item.href)) {
    return "div";
  }

  return segment.item.to ? "router-link" : "a";
}

function getSegmentBindings(segment: ItemSegment) {
  if (segment.isCurrent || (!segment.item.to && !segment.item.href)) {
    return {};
  }

  if (segment.item.to) {
    return { to: segment.item.to };
  }

  return {
    href: segment.item.href,
    target: "_blank",
    rel: "noopener noreferrer",
  };
}
</script>

<style>
.breadcrumbs {
  display: flex;
  min-width: 0;
  align-items: center;
  overflow: hidden;
  white-space: nowrap;
  color: var(--text-color-muted);
}

.breadcrumbs-base {
  min-height: 3rem;
  gap: 0.25rem;
  font-size: 1rem;
}

.breadcrumbs-sm {
  min-height: 2.5rem;
  gap: 0.125rem;
  font-size: 0.875rem;
}

.breadcrumbs-mobile {
  display: contents;
}

.breadcrumbs-desktop {
  display: none;
}

.breadcrumbs-link,
.breadcrumbs-current {
  display: inline-flex;
  min-width: 0;
  max-width: 100%;
  align-items: center;
  gap: 0.25rem;
}

.breadcrumbs-link {
  color: var(--text-color-muted);
  text-decoration: none;
  transition:
    color 0.2s ease,
    opacity 0.2s ease;

  &:hover {
    color: var(--text-color);
  }
}

.breadcrumbs-current {
  color: var(--text-color);
  font-weight: 500;
}

.breadcrumbs-icon {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
}

.breadcrumbs-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.breadcrumbs-separator,
.breadcrumbs-ellipsis {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  color: color-mix(in srgb, var(--text-color-muted) 80%, transparent);
}

.breadcrumbs-separator svg {
  width: 1.5rem;
  height: 1.5rem;
}

@media all and (min-width: 640px) {
  .breadcrumbs-mobile {
    display: none;
  }

  .breadcrumbs-desktop {
    display: contents;
  }
}
</style>
