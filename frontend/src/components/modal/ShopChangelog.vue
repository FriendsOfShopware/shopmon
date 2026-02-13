<template>
  <modal :show="show" close-x-mark @close="$emit('close')">
    <template #title>
      Shop changelog -
      <span v-if="changelog?.date" class="modal-changelog-title-date">{{
        formatDateTime(changelog.date)
      }}</span>
    </template>

    <template #content>
      <template v-if="changelog?.oldShopwareVersion && changelog?.newShopwareVersion">
        Shopware update from <strong>{{ changelog.oldShopwareVersion }}</strong> to
        <strong>{{ changelog.newShopwareVersion }}</strong>
      </template>

      <div v-if="changelog?.extensions?.length">
        <h2 class="modal-changelog-subtitle">Shop Plugin Changelog:</h2>

        <ul class="modal-changelog-logs">
          <li v-for="extension in changelog?.extensions" :key="extension.name">
            <div class="modal-changelog-extension-name">
              {{ extension.label }} <span>({{ extension.name }})</span>
            </div>

            {{ extension.state }}
            <template v-if="extension.state === 'installed' && extension.active">
              and activated
            </template>
            <template v-if="extension.state === 'updated'">
              from {{ extension.old_version }} to {{ extension.new_version }}
            </template>
            <template v-else>
              {{ extension.new_version }}
              <template v-if="!extension.new_version">
                {{ extension.old_version }}
              </template>
            </template>
          </li>
        </ul>
      </div>
    </template>
  </modal>
</template>

<script setup lang="ts">
import { formatDateTime } from "@/helpers/formatter";
import type { ShopChangelog } from "@/types/shop";

interface Props {
  show: boolean;
  changelog: ShopChangelog | null;
}

defineProps<Props>();
defineEmits<{
  close: [];
}>();
</script>

<style scoped>
.modal-changelog-title-date {
  font-weight: normal;
  color: var(--text-color-muted);
}

.modal-changelog-subtitle {
  font-size: 1.1rem;
  margin-bottom: 0.5rem;
  margin-top: 0.5rem;
}

.modal-changelog-logs {
  list-style: disc;

  li {
    margin-left: 1rem;
    margin-bottom: 0.5rem;
  }
}

.modal-changelog-extension-name {
  font-weight: 500;

  span {
    font-weight: normal;
    color: var(--text-color-muted);
  }
}
</style>
