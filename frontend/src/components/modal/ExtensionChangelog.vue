<template>
  <modal :show="show" close-x-mark @close="$emit('close')">
    <template #title>
      {{ $t("extensionChangelog.title") }} -
      <span class="extension-changelog-name">{{ extension?.name }}</span>
    </template>

    <template #content>
      <ul v-if="changelogEntries.length > 0" class="extension-changelog">
        <li
          v-for="changeLog in changelogEntries"
          :key="changeLog.version"
          class="extension-changelog-item"
        >
          <div class="extension-changelog-title">
            <span
              v-if="!changeLog.isCompatible"
              :data-tooltip="$t('extensionChangelog.notCompatible')"
            >
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

      <alert v-else type="error"> {{ $t("extensionChangelog.noData") }} </alert>
    </template>
  </modal>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { formatDate } from "@/helpers/formatter";
import type { ExtensionWithChangelog } from "@/composables/useExtensionChangelogModal";

interface Props {
  show: boolean;
  extension: ExtensionWithChangelog | null;
}

const props = defineProps<Props>();
defineEmits<{
  close: [];
}>();

const changelogEntries = computed(() => {
  const cl = props.extension?.changelog;
  if (Array.isArray(cl)) {
    return cl;
  }
  return [];
});
</script>

<style scoped>
.extension-changelog {
  list-style: none;
  margin: 0;
  padding: 0;
}

.extension-changelog-name {
  font-weight: normal;
  color: var(--text-color-muted);
}

.extension-changelog-item {
  margin-bottom: 1.5rem;

  &:last-child {
    margin-bottom: 0;
  }
}

.extension-changelog-title {
  font-weight: 600;
  margin-bottom: 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.extension-changelog-date {
  font-weight: normal;
  color: var(--text-color-muted);
}

.extension-changelog-content {
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
</style>
