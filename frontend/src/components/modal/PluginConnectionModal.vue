<template>
  <modal :show="show" close-x-mark @close="$emit('close')">
    <template #title> {{ $t("pluginModal.title") }} </template>
    <template #content>
      <p>{{ $t("pluginModal.description") }}</p>
      <textarea
        v-model="pluginBase64Value"
        class="field"
        rows="4"
        placeholder="eyJ1cmwiOiJodHRwczpcL1wvZGVtby5mb3MuZ2ciLCJjbGllbnRJZCI6..."
        @input="$emit('update:base64', pluginBase64Value)"
      />
      <div v-if="error" class="field-error-message">
        {{ error }}
      </div>
    </template>
    <template #footer>
      <button type="button" class="btn btn-primary" @click="$emit('import')">
        {{ $t("pluginModal.importData") }}
      </button>
      <button type="button" class="btn btn-cancel" @click="$emit('close')">
        {{ $t("common.cancel") }}
      </button>
    </template>
  </modal>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

interface Props {
  show: boolean;
  base64: string;
  error?: string;
}

const props = withDefaults(defineProps<Props>(), {
  error: "",
});

defineEmits<{
  close: [];
  import: [];
  "update:base64": [value: string];
}>();

const pluginBase64Value = ref(props.base64);

// Watch for external base64 changes
watch(
  () => props.base64,
  (newValue) => {
    pluginBase64Value.value = newValue;
  },
);

// Reset when modal is closed
watch(
  () => props.show,
  (newShow) => {
    if (!newShow) {
      pluginBase64Value.value = "";
    }
  },
);
</script>
