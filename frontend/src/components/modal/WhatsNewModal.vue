<template>
  <modal :show="show" close-x-mark @close="$emit('close')">
    <template #icon>
      <div class="whats-new-icon">
        <icon-fa6-solid:rocket class="icon" />
      </div>
    </template>

    <template #title>What's new: Packages Mirror</template>

    <template #content>
      <div class="whats-new-hero">
        <span class="whats-new-badge">New in March 2026</span>
        <h4>75x faster Shopware store packages via Global CDN</h4>
        <p>
          Add your Shopware store tokens to a project and serve packages through the Shopmon
          packages mirror — ~80ms instead of ~6s from
          <code>packages.shopware.com</code>.
        </p>
      </div>

      <ul class="whats-new-list">
        <li>
          <icon-fa6-solid:bolt class="icon" />
          Packages are cached on a Global CDN for near-instant Composer installs.
        </li>
        <li>
          <icon-fa6-solid:arrows-rotate class="icon" />
          Tokens sync automatically every hour — or trigger a manual sync any time.
        </li>
        <li>
          <icon-fa6-solid:shield-halved class="icon" />
          Tokens are validated against the Shopware store before being saved.
        </li>
      </ul>

      <div class="whats-new-setup">
        <p class="whats-new-setup-title">Quick setup</p>
        <p>
          Add a Shopware store token in your project settings, then replace
          <code>packages.shopware.com</code> with the mirror URL in your <code>composer.json</code>.
          The exact snippets are shown after adding a token.
        </p>
      </div>

      <div v-if="sponsors.length" class="whats-new-sponsors">
        <p class="whats-new-setup-title">Sponsors</p>
        <p class="whats-new-sponsors-copy">
          The public start page highlights the companies supporting ongoing Shopmon development.
        </p>

        <sponsor-showcase :sponsors="sponsors" compact />
      </div>
    </template>

    <template #footer>
      <router-link :to="{ name: 'account.project.list' }" class="btn" @click="$emit('close')">
        <icon-fa6-solid:folder-open class="icon" />
        Open Projects
      </router-link>

      <router-link :to="{ name: 'account.docs' }" class="btn" @click="$emit('close')">
        <icon-fa6-solid:book class="icon" />
        Documentation
      </router-link>

      <button type="button" class="btn btn-primary" @click="$emit('close')">Close</button>
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
    color: var(--text-color);
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
  color: var(--text-color);

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
  color: var(--text-color);
}
</style>
