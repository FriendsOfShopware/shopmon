<template>
  <modal :show="show" close-x-mark @close="$emit('close')">
    <template #title>
      Changelog - <span class="extension-changelog-name">{{ extension?.name }}</span>
    </template>

    <template #content>
      <ul v-if="extension?.changelog?.length > 0" class="extension-changelog">
        <li
          v-for="changeLog in extension.changelog"
          :key="changeLog.version"
          class="extension-changelog-item"
        >
          <div class="extension-changelog-title">
            <span v-if="!changeLog.isCompatible" data-tooltip="not compatible with your version">
              <icon-fa6-solid:circle-info class="icon icon-warning" />
            </span>

            {{ changeLog.version }} -
            <span class="extension-changelog-date">
              {{ formatDate(changeLog.creationDate) }}
            </span>
          </div>

          <!-- eslint-disable vue/no-v-html -->
          <div class="extension-changelog-content" v-html="changeLog.text" />
          <!-- eslint-enable vue/no-v-html -->
        </li>
      </ul>

      <alert v-else type="error"> No Changelog data provided </alert>
    </template>
  </modal>
</template>

<script setup lang="ts">
import { formatDate } from "@/helpers/formatter";

interface ExtensionChangelog {
  version: string;
  text: string;
  creationDate: string;
  isCompatible: boolean;
}

interface Extension {
  name: string;
  label: string;
  changelog: ExtensionChangelog[] | null;
}

interface Props {
  show: boolean;
  extension: Extension | null;
}

defineProps<Props>();
defineEmits<{
  close: [];
}>();
</script>

<style scoped>
.extension-changelog {
  list-style: none;
  margin: 0;
  padding: 0;

  &-name {
    font-weight: normal;
    color: var(--text-color-muted);
  }

  &-item {
    margin-bottom: 1.5rem;

    &:last-child {
      margin-bottom: 0;
    }
  }

  &-title {
    font-weight: 600;
    margin-bottom: 0.5rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  &-date {
    font-weight: normal;
    color: var(--text-color-muted);
  }

  &-content {
    line-height: 1.6;

    :deep(ul) {
      margin: 0.5rem 0;
      padding-left: 1.5rem;
    }

    :deep(li) {
      margin: 0.25rem 0;
    }

    :deep(p) {
      margin: 0.5rem 0;
    }

    :deep(strong) {
      font-weight: 600;
    }
  }
}
</style>
