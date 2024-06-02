<template>
    <transition-root
        as="template"
        :show="show"
    >
        <headless-dialog
            as="div"
            class="modal"
            @close="emit('close')"
        >
            <transition-child
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div
                    class="modal-overlay transform translate"
                />
            </transition-child>

            <div class="modal-container">
                <div class="modal-panel-wrapper">
                    <transition-child
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0 scale-80"
                        enter-to="opacity-100 scale-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100 scale-100"
                        leave-to="opacity-0 scale-80"
                    >
                        <dialog-panel
                            class="modal-panel transform transition"
                        >
                            <button
                                v-if="closeXMark"
                                class="modal-close-button"
                                type="button"
                                @click="emit('close')"
                            >
                                <icon-fa6-solid:xmark
                                    aria-hidden="true"
                                    size="xl"
                                />
                            </button>

                            <div class="modal-grid">
                                <div v-if="!!$slots.icon" class="modal-icon">
                                    <slot name="icon" />
                                </div>

                                <div
                                    class="modal-content"
                                >
                                    <dialog-title
                                        v-if="!!$slots.title"
                                        as="h3"
                                        class="modal-title"
                                    >
                                        <slot name="title" />
                                    </dialog-title>

                                    <slot name="content" />
                                </div>
                            </div>

                            <div
                                v-if="!!$slots.footer"
                                class="modal-footer"
                            >
                                <slot name="footer" />
                            </div>
                        </dialog-panel>
                    </transition-child>
                </div>
            </div>
        </headless-dialog>
    </transition-root>
</template>

<script setup lang="ts">
import { Dialog as HeadlessDialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';

defineProps<{ show: boolean, closeXMark?: boolean }>();

const emit = defineEmits<{ (e: 'close'): void }>();
</script>

<style>
.modal {
    position: relative;
    z-index: 10;
}

.modal-overlay {
    position: fixed;
    inset: 0;
    background-color: #6b7280bf;
    transition: opacity 0.3s ease;

    .dark & {
        background-color: #171717cc;
    }
}

.modal-container {
    position: fixed;
    inset: 0;
    overflow-y: auto;
}

.modal-panel-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100%;
    padding: 1rem;
    text-align: center;

    @media (min-width: 640px) {
        padding: 0;
    }
}

.modal-panel {
    position: relative;
    background-color: var(--panel-background);
    border-radius: 0.5rem;
    padding: 1rem;
    padding-top: 1.25rem;
    padding-bottom: 1rem;
    text-align: left;
    overflow: hidden;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
    max-width: 32rem;
    width:90%;

    @media (min-width: 640px) {
        padding: 1.5rem;
    }
}

.modal-close-button {
    position: absolute;
    top: 0.25rem;
    right: 0.625rem;
    font-size: 1.25rem;
    background: none;
    border: none;
    outline: none;
    cursor: pointer;
    transition: opacity .4s;
    opacity: .5;

    &:hover {
        opacity: 1;
    }
}

.modal-grid {
    display: flex;
    align-items: flex-start;
}

.modal-icon {
    .icon {
        width: 1.5rem;
        height: 1.5rem;
    }

    + .modal-content {
        margin-left: 1rem;
    }
}

.modal-content {
    text-align: left;
    flex-grow: 1;
}

.modal-title {
    font-size: 1.125rem;
    line-height: 1.5rem;
    font-weight: 500;
    margin-bottom: 0.5rem;
}

.modal-footer {
    margin-top: 1.25rem;
    display: flex;
    justify-content: flex-end;
    flex-direction: column;
    gap: 1rem;

    @media (min-width: 640px) {
        flex-direction: row;
    }

    .btn-cancel {
        @media (min-width: 640px) {
            order: -1
        }
    }

    .btn {
        display: block;
    }
}
</style>
