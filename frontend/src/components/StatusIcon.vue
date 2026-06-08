<template>
  <span v-if="tooltip" :title="status" class="capitalize">
    <component :is="getIconComponent(status)" class="size-4" :class="getIconClasses(status)" />
  </span>

  <template v-else>
    <component :is="getIconComponent(status)" class="size-4" :class="getIconClasses(status)" />
  </template>
</template>

<script setup>
import FaCircle from "~icons/fa6-regular/circle";
import FaCircleCheck from "~icons/fa6-solid/circle-check";
import FaCircleInfo from "~icons/fa6-solid/circle-info";
import FaCircleXmark from "~icons/fa6-solid/circle-xmark";

defineProps({
  status: {
    type: String,
    required: true,
  },
  tooltip: {
    type: Boolean,
    required: false,
    default: false,
  },
});

function getIconComponent(status) {
  switch (status) {
    case "red":
    case "inactive":
      return FaCircleXmark;
    case "yellow":
      return FaCircleInfo;
    case "not installed":
      return FaCircle;
    default:
      return FaCircleCheck;
  }
}

function getIconClasses(status) {
  switch (status) {
    case "red":
      return "text-destructive";
    case "yellow":
      return "text-warning";
    case "inactive":
    case "not installed":
      return "text-muted-foreground";
    default:
      return "text-success";
  }
}
</script>
