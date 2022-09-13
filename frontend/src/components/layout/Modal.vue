<template>
    <TransitionRoot as="template" :show="show">
      <Dialog as="div" class="relative z-10" @close="emit('close')">
        <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100"
          leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
          <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </TransitionChild>

        <div class="fixed z-10 inset-0 overflow-y-auto">
          <div class="flex items-center justify-center min-h-full p-4 text-center sm:p-0">
            <TransitionChild as="template" enter="ease-out duration-300"
              enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200"
              leave-from="opacity-100 translate-y-0 sm:scale-100"
              leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
              <DialogPanel
                class="relative bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:max-w-lg sm:w-full sm:p-6">
                <button class="absolute top-1 right-2.5 focus:outline-none text-base" @click="emit('close')" v-if="closeXMark">
                  <icon-fa6-solid:xmark aria-hidden="true" size="xl"/>
                </button>

                <div class="flex items-start">
                    <div v-if="!!$slots.icon">                        
                        <slot name="icon"></slot>
                    </div>                        
                    <div class="text-left" :class="[{'ml-4': !!$slots.icon}]">
                        <DialogTitle as="h3" class="text-lg leading-6 font-medium text-gray-900" v-if="!!$slots.title">
                            <slot name="title"></slot>
                        </DialogTitle>
                        
                        <slot name="content"></slot>
                    </div>
                </div>
                    
                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse" v-if="!!$slots.footer">
                  <slot name="footer"></slot>
                </div>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </TransitionRoot>
</template>

<script setup lang="ts">
    import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';

    defineProps<{show: boolean, closeXMark?: boolean}>();

    const emit = defineEmits<{(e: 'close'): void}>();

</script>