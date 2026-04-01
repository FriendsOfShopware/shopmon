<template>
  <modal :show="show" close-x-mark @close="$emit('close')">
    <template #icon>
      <div class="whats-new-icon">
        <icon-fa6-solid:rocket class="icon" />
      </div>
    </template>

    <template #title>{{ $t('whatsNew.title') }}</template>

    <template #content>
      <div class="whats-new-hero">
        <span class="whats-new-badge">{{ $t('whatsNew.badge') }}</span>
        <h4>{{ $t('whatsNew.heroTitle') }}</h4>
        <p>
          {{ $t('whatsNew.heroDesc') }}
        </p>
      </div>

      <ul class="whats-new-list">
        <li>
          <icon-fa6-solid:bolt class="icon" />
          {{ $t('whatsNew.featureCdn') }}
        </li>
        <li>
          <icon-fa6-solid:arrows-rotate class="icon" />
          {{ $t('whatsNew.featureSync') }}
        </li>
        <li>
          <icon-fa6-solid:shield-halved class="icon" />
          {{ $t('whatsNew.featureValidation') }}
        </li>
      </ul>

      <div class="whats-new-setup">
        <p class="whats-new-setup-title">{{ $t('whatsNew.quickSetup') }}</p>
        <p>
          {{ $t('whatsNew.quickSetupDesc') }}
        </p>
      </div>

      <div v-if="sponsors.length" class="whats-new-sponsors">
        <p class="whats-new-setup-title">{{ $t('whatsNew.sponsors') }}</p>
        <p class="whats-new-sponsors-copy">
          {{ $t('whatsNew.sponsorsDesc') }}
        </p>

        <sponsor-showcase :sponsors="sponsors" compact />
      </div>
    </template>

    <template #footer>
      <router-link :to="{ name: 'account.project.list' }" class="btn" @click="$emit('close')">
        <icon-fa6-solid:folder-open class="icon" />
        {{ $t('whatsNew.openProjects') }}
      </router-link>

      <router-link :to="{ name: 'account.docs' }" class="btn" @click="$emit('close')">
        <icon-fa6-solid:book class="icon" />
        {{ $t('nav.documentation') }}
      </router-link>

      <button type="button" class="btn btn-primary" @click="$emit('close')">{{ $t('whatsNew.close') }}</button>
    </template>
  </modal>
</template>

<script setup lang="ts">
import Modal from "@/components/layout/Modal.vue";
import SponsorShowcase from "@/components/SponsorShowcase.vue";
import { sponsors } from "@/data/sponsors";

defineProps<{ show: boolean }>();

defineEmits<{ close: [] }>();
</script>

<style scoped>
.whats-new-icon {
  width: 3rem;
  height: 3rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  color: #fff;
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--primary-color) 92%, white 8%),
    #0369a1
  );
}

.whats-new-hero {
  margin-bottom: 1rem;

  h4 {
    margin: 0.75rem 0 0.5rem;
    font-size: 1.25rem;
    color: var(--item-title-color);
  }

  p {
    margin: 0;
    color: var(--text-color-muted);
  }
}

.whats-new-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--primary-color) 14%, transparent);
  color: var(--link-color);
  font-size: 0.8125rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.whats-new-list {
  list-style: none;
  padding: 0;
  margin: 0 0 1rem;
  display: grid;
  gap: 0.75rem;

  li {
    display: grid;
    grid-template-columns: 1.25rem 1fr;
    gap: 0.75rem;
    padding: 0.85rem 1rem;
    border: 1px solid var(--panel-border-color);
    border-radius: 0.75rem;
    background: var(--item-background);
  }

  .icon {
    width: 1rem;
    height: 1rem;
    margin-top: 0.2rem;
    color: var(--primary-color);
  }
}

.whats-new-setup {
  padding: 1rem;
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--primary-color) 8%, var(--panel-background));
  border: 1px solid color-mix(in srgb, var(--primary-color) 18%, var(--panel-border-color));

  p {
    margin: 0;
  }
}

.whats-new-setup-title {
  font-weight: 700;
  color: var(--item-title-color);
  margin-bottom: 0.35rem !important;
}

.whats-new-sponsors {
  margin-top: 1rem;
}

.whats-new-sponsors-copy {
  margin: 0 0 1rem;
}
</style>
