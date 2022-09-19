<script lang="ts" setup>
defineProps<{title: string}>()
</script>

<template>
  <header class="bg-white shadow dark:bg-neutral-800 dark:bg-opacity-50 dark:border-b dark:border-neutral-700">
    <div class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between">
        <h1 class="text-3xl tracking-tight font-bold leading-tight">
          {{ title }}
        </h1>
        <slot/>
      </div> 
    </div>
  </header>
</template>