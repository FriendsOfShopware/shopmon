<script lang="ts" setup>
import { Slot } from "reka-ui";
import { useFormField } from "./useFormField";

const { error, formItemId, formDescriptionId, formMessageId } = useFormField();
</script>

<template>
  <!-- @vue-ignore Slot merges id and aria attributes onto its child -->
  <Slot
    :id="formItemId"
    data-slot="form-control"
    :aria-describedby="!error ? `${formDescriptionId}` : `${formDescriptionId} ${formMessageId}`"
    :aria-invalid="!!error"
  >
    <slot />
  </Slot>
</template>
