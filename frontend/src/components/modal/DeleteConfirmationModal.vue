<template>
  <modal :show="show" close-x-mark class="modal-confirm" @close="$emit('close')">
    <template #icon>
      <icon-fa6-solid:triangle-exclamation class="icon icon-error" aria-hidden="true" />
    </template>

    <template #title>
      {{ title }}
    </template>

    <template #content>
      <div v-if="customMessage">
        {{ customMessage }}
      </div>

      <div v-else>
        <p>
          Are you sure you want to delete <strong>{{ entityName }}</strong
          >?
        </p>

        <p>
          All of your data will be permanently removed from our servers forever.
          <strong>This action cannot be undone.</strong>
        </p>

        <p v-if="customConsequence" class="text-muted">
          {{ customConsequence }}
        </p>
      </div>

      <!-- Password confirmation field -->
      <div v-if="requirePassword" class="section-password">
        <label for="deletePassword">Current Password</label>
        <input
          id="deletePassword"
          v-model="passwordValue"
          type="password"
          class="field"
          autocomplete="off"
          @input="$emit('password-change', passwordValue)"
        />
      </div>
    </template>

    <template #footer>
      <button
        type="button"
        class="btn btn-danger"
        :disabled="isLoading || (requirePassword && !passwordValue)"
        @click="$emit('confirm')"
      >
        <icon-fa6-solid:trash v-if="!isLoading" class="icon" />
        <icon-line-md:loading-twotone-loop v-if="isLoading" class="icon" />
        {{ confirmButtonText }}
      </button>

      <button type="button" class="btn btn-cancel" @click="$emit('close')">Cancel</button>
    </template>
  </modal>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

interface Props {
  show: boolean;
  title: string;
  entityName: string;
  customMessage?: string;
  customConsequence?: string;
  requirePassword?: boolean;
  isLoading?: boolean;
  confirmButtonText?: string;
  reversedButtons?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  customMessage: "",
  customConsequence: "",
  requirePassword: false,
  isLoading: false,
  confirmButtonText: "Delete",
  reversedButtons: false,
});

defineEmits<{
  close: [];
  confirm: [];
  "password-change": [password: string];
}>();

const passwordValue = ref("");

// Reset password when modal is closed
watch(
  () => props.show,
  (newShow) => {
    if (!newShow) {
      passwordValue.value = "";
    }
  },
);
</script>

<style>
.modal-confirm {
  .modal-panel {
    max-width: 32rem;
  }
}
</style>

<style scoped>
label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.text-muted {
  color: var(--text-color-muted);
}

.modal-footer-reversed {
  display: flex;
  flex-direction: row-reverse;
  gap: 0.5rem;
}

.section-password {
  margin-top: 1rem;
}
</style>
