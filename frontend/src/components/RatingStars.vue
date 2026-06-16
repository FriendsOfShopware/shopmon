<template>
  <span
    v-if="rating !== null"
    :title="$t('common.ratingOutOfFive', { rating: rating / 2 })"
    class="inline-flex whitespace-nowrap"
  >
    <template v-for="n in 5" :key="n">
      <icon-fa6-regular:star v-if="rating / 2 - n < -0.5" />
      <icon-fa6-solid:star-half-stroke v-else-if="rating / 2 - n === -0.5" />
      <icon-fa6-solid:star v-else />
    </template>
  </span>
</template>

<script lang="ts" setup>
defineProps<{ rating: number | null }>();
</script>
