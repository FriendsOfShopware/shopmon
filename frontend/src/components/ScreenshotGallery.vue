<template>
  <div>
    <!-- Main slide -->
    <div
      ref="frame"
      class="group relative aspect-[16/10] w-full overflow-hidden rounded-xl border bg-muted shadow-sm"
    >
      <button
        type="button"
        class="block size-full focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        :aria-label="$t('extensions.openScreenshot')"
        @click="openLightbox(active)"
      >
        <img
          :key="images[active]"
          :src="images[active]"
          :alt="`${alt} ${active + 1}`"
          class="size-full object-cover"
        />
      </button>

      <template v-if="images.length > 1">
        <button
          type="button"
          class="absolute left-2 top-1/2 flex size-8 -translate-y-1/2 items-center justify-center rounded-full bg-background/80 text-foreground opacity-0 shadow-sm transition hover:bg-background focus:opacity-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring group-hover:opacity-100"
          :aria-label="$t('common.previous')"
          @click="prev"
        >
          <icon-fa6-solid:chevron-left class="size-3" />
        </button>
        <button
          type="button"
          class="absolute right-2 top-1/2 flex size-8 -translate-y-1/2 items-center justify-center rounded-full bg-background/80 text-foreground opacity-0 shadow-sm transition hover:bg-background focus:opacity-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring group-hover:opacity-100"
          :aria-label="$t('common.next')"
          @click="next"
        >
          <icon-fa6-solid:chevron-right class="size-3" />
        </button>
        <span
          class="absolute bottom-2 right-2 rounded-full bg-background/80 px-2 py-0.5 text-xs font-medium tabular-nums text-foreground"
        >
          {{ active + 1 }} / {{ images.length }}
        </span>
      </template>
    </div>

    <!-- Thumbnails -->
    <div v-if="images.length > 1" class="mt-3 flex flex-wrap gap-2">
      <button
        v-for="(url, i) in images"
        :key="url"
        type="button"
        class="h-14 w-20 shrink-0 overflow-hidden rounded-lg border transition focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        :class="
          i === active
            ? 'border-primary ring-2 ring-primary/30'
            : 'border-border opacity-70 hover:opacity-100'
        "
        :aria-label="`${alt} ${i + 1}`"
        @click="active = i"
      >
        <img :src="url" alt="" class="size-full object-cover" />
      </button>
    </div>

    <!-- Lightbox -->
    <Teleport to="body">
      <div
        v-if="lightboxOpen"
        ref="lightboxEl"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        @click="lightboxOpen = false"
      >
        <button
          type="button"
          class="absolute right-4 top-4 flex size-10 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
          :aria-label="$t('common.close')"
          @click="lightboxOpen = false"
        >
          <icon-fa6-solid:xmark class="size-4" />
        </button>

        <button
          v-if="images.length > 1"
          type="button"
          class="absolute left-4 flex size-11 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
          :aria-label="$t('common.previous')"
          @click.stop="prev"
        >
          <icon-fa6-solid:chevron-left class="size-5" />
        </button>

        <img
          :src="images[active]"
          :alt="`${alt} ${active + 1}`"
          class="max-h-full max-w-5xl rounded-xl object-contain shadow-2xl"
          @click.stop
        />

        <button
          v-if="images.length > 1"
          type="button"
          class="absolute right-4 flex size-11 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
          :aria-label="$t('common.next')"
          @click.stop="next"
        >
          <icon-fa6-solid:chevron-right class="size-5" />
        </button>

        <span
          v-if="images.length > 1"
          class="absolute bottom-5 left-1/2 -translate-x-1/2 rounded-full bg-black/50 px-3 py-1 text-sm text-white tabular-nums"
        >
          {{ active + 1 }} / {{ images.length }}
        </span>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { onKeyStroke, useSwipe } from "@vueuse/core";

const props = withDefaults(
  defineProps<{
    images: string[];
    alt?: string;
  }>(),
  { alt: "" },
);

const active = ref(0);
const lightboxOpen = ref(false);
const frame = ref<HTMLElement | null>(null);
const lightboxEl = ref<HTMLElement | null>(null);

// Reset to the first slide whenever the image set changes.
watch(
  () => props.images,
  () => {
    active.value = 0;
  },
);

function next() {
  if (props.images.length === 0) return;
  active.value = (active.value + 1) % props.images.length;
}

function prev() {
  if (props.images.length === 0) return;
  active.value = (active.value - 1 + props.images.length) % props.images.length;
}

function openLightbox(i: number) {
  active.value = i;
  lightboxOpen.value = true;
}

// Arrow keys page through the slides whenever the lightbox is open; Escape closes it.
onKeyStroke("ArrowLeft", () => lightboxOpen.value && prev());
onKeyStroke("ArrowRight", () => lightboxOpen.value && next());
onKeyStroke("Escape", () => {
  lightboxOpen.value = false;
});

// Touch swipe on the inline frame and inside the lightbox.
useSwipe(frame, {
  onSwipeEnd(_e, direction) {
    if (direction === "left") next();
    else if (direction === "right") prev();
  },
});
useSwipe(lightboxEl, {
  onSwipeEnd(_e, direction) {
    if (direction === "left") next();
    else if (direction === "right") prev();
  },
});
</script>
