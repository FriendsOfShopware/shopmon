<script setup lang="ts">
import Logo from '@/components/layout/Logo.vue';
</script>

<template>
    <div class="
      min-h-full
      flex
      items-center
      justify-center
      py-12
      px-4
      sm:px-6
      lg:px-8
    ">
        <div class="max-w-md w-full space-y-8">
            <Logo />
            <slot />
        </div>
    </div>
</template>