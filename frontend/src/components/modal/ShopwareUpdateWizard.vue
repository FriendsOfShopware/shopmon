<template>
  <modal class="update-wizard" :show="show" close-x-mark @close="$emit('close')">
    <template #title>
      <icon-fa6-solid:rotate />
      Shopware Extension Compatibility Check
    </template>

    <template #content>
      <select
        class="field"
        @change="
          (event: Event) => $emit('versionSelected', (event.target as HTMLSelectElement).value)
        "
      >
        <option disabled selected>Select update Version</option>

        <option v-for="version in shopwareVersions" :key="version">
          {{ version }}
        </option>
      </select>

      <template v-if="loading">
        <div class="update-wizard-loader">
          Loading
          <icon-fa6-solid:rotate class="animate-spin" />
        </div>
      </template>

      <div v-if="extensions" :class="{ 'update-wizard-refresh': loading }">
        <h2 class="update-wizard-plugins-heading">Extension Compatibility</h2>

        <ul>
          <li v-for="extension in extensions" :key="extension.name" class="update-wizard-plugin">
            <div class="update-wizard-plugin-icon">
              <icon-fa6-regular:circle v-if="!extension.active" class="icon icon-muted" />
              <icon-fa6-solid:circle-info
                v-else-if="!extension.compatibility"
                class="icon icon-warning"
              />
              <icon-fa6-solid:circle-xmark
                v-else-if="extension.compatibility.type == 'red'"
                class="icon icon-error"
              />
              <icon-fa6-solid:rotate
                v-else-if="extension.compatibility.label === 'Available now'"
                class="icon icon-info"
              />
              <icon-fa6-solid:circle-check v-else class="icon icon-success" />
            </div>

            <div>
              <component
                :is="extension.storeLink ? 'a' : 'span'"
                v-bind="extension.storeLink ? { href: extension.storeLink, target: '_blank' } : {}"
              >
                <strong>{{ extension.label }}</strong>
              </component>
              <span class="update-wizard-plugin-technical-name"> ({{ extension.name }})</span>

              <div v-if="!extension.compatibility || !extension.storeLink">
                This plugin is not available in the Store. Please contact the plugin manufacturer.
              </div>
              <div v-else>
                {{ extension.compatibility.label }}
              </div>
            </div>
          </li>
        </ul>
      </div>
    </template>
  </modal>
</template>

<script setup lang="ts">
interface Props {
  show: boolean;
  shopwareVersions: string[] | null;
  loading: boolean;
  extensions: any[] | null;
}

defineProps<Props>();
defineEmits<{
  close: [];
  versionSelected: [version: string];
}>();
</script>

<style scoped>
.update-wizard {
  .field {
    margin-bottom: 0.75rem;
  }

  &-loader {
    text-align: center;
  }

  &-refresh {
    opacity: 0.2;
  }

  &-plugins {
    &-heading {
      font-size: 1.125rem;
      font-weight: 500;
      margin-bottom: 0.5rem;
    }
  }

  &-plugin {
    background-color: var(--item-background);
    padding: 0.5rem;
    display: flex;

    &:nth-child(odd) {
      background-color: var(--item-odd-background);
    }

    &:hover {
      background-color: var(--item-hover-background);
    }

    &-icon {
      margin-right: 0.5rem;
    }

    &-technical-name {
      opacity: 0.6;
    }
  }
}
</style>
