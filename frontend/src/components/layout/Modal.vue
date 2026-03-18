<template>
  <dialog
    ref="dialogRef"
    class="modal-dialog"
    @close="emit('close')"
    @click="onBackdropClick"
  >
    <div class="modal-panel">
      <button
        v-if="closeXMark"
        class="modal-close-button"
        type="button"
        @click="emit('close')"
      >
        <icon-fa6-solid:xmark aria-hidden="true" size="xl" />
      </button>

      <div class="modal-grid">
        <div v-if="!!$slots.icon" class="modal-icon">
          <slot name="icon" />
        </div>

        <div class="modal-content">
          <h3 v-if="!!$slots.title" class="modal-title">
            <slot name="title" />
          </h3>

          <slot name="content" />
        </div>
      </div>

      <div v-if="!!$slots.footer" class="modal-footer">
        <slot name="footer" />
      </div>
    </div>
  </dialog>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from "vue";

const props = defineProps<{ show: boolean; closeXMark?: boolean }>();

const emit = defineEmits<{ close: [] }>();

const dialogRef = ref<HTMLDialogElement | null>(null);

function syncDialog(show: boolean) {
  const el = dialogRef.value;
  if (!el) return;

  if (show && !el.open) {
    el.showModal();
  } else if (!show && el.open) {
    el.close();
  }
}

watch(() => props.show, syncDialog);

onMounted(() => syncDialog(props.show));

function onBackdropClick(event: MouseEvent) {
  if (event.target === dialogRef.value) {
    emit("close");
  }
}
</script>

<style>
.modal-dialog {
  border: none;
  padding: 0;
  background: transparent;
  max-width: none;
  max-height: none;
  overflow: visible;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-dialog:not([open]) {
  display: none;
}

.modal-dialog::backdrop {
  background-color: #6b7280bf;
}

.dark .modal-dialog::backdrop,
:global(.dark) .modal-dialog::backdrop {
  background-color: #171717cc;
}

.modal-panel {
  position: relative;
  background-color: var(--panel-background);
  border-radius: 0.5rem;
  padding: 1.25rem 1rem 1rem;
  text-align: left;
  overflow: hidden;
  box-shadow:
    0 10px 15px -3px rgba(0, 0, 0, 0.1),
    0 4px 6px -2px rgba(0, 0, 0, 0.05);
  max-width: 52rem;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;

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
  transition: opacity 0.4s;
  opacity: 0.5;

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
      order: -1;
    }
  }

  .btn {
    display: block;
  }
}
</style>
