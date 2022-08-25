<script setup>
import { storeToRefs } from 'pinia';

import { useAlertStore } from '@/stores';
import { XIcon } from '@heroicons/vue/solid';

const alertStore = useAlertStore();
const { alert } = storeToRefs(alertStore);
</script>

<style>
.alert-danger {
  @apply text-red-900 border-red-200 bg-red-50;
}

.alert-success {
  @apply text-green-900 border-green-200 bg-green-50;
}
</style>

<template>
  <transition
    enter-active-class="transition ease-out duration-200"
    enter-from-class="opacity-0 -translate-y-1"
    enter-to-class="opacity-100 translate-y-0"
    leave-active-class="transition ease-in duration-150"
    leave-from-class="opacity-100 translate-y-0"
    leave-to-class="opacity-0 -translate-y-1"
  >
    <div v-if="alert" class="w-full absolute flex justify-center p-3">
      <div
        class="relative max-w-xl w-full rounded p-3 border inset-0"
        :class="alert.type"
      >
        <button
          @click="alertStore.clear()"
          class="w-4 h-4 absolute top-1 right-1 color-red-500"
        >
          <XIcon />
        </button>
        {{ alert.message }}
      </div>
    </div>
  </transition>
</template>
