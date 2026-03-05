<template>
  <modal :show="show" close-x-mark @close="$emit('close')">
    <template #icon>
      <div class="whats-new-icon">
        <icon-fa6-solid:rocket class="icon" />
      </div>
    </template>

    <template #title>What's new: Deployment tracking</template>

    <template #content>
      <div class="whats-new-hero">
        <span class="whats-new-badge">New in March 2026</span>
        <h4>Track deployments directly inside each shop</h4>
        <p>
          You can report deployments with <code>shopmon-cli</code> and review them later from the
          <strong>Deployments</strong> section of a shop.
        </p>
      </div>

      <ul class="whats-new-list">
        <li>
          <icon-fa6-solid:clock-rotate-left class="icon" />
          See a full deployment history with timestamps, duration, and success or failure state.
        </li>
        <li>
          <icon-fa6-solid:file-lines class="icon" />
          Open each deployment to inspect command output, exit code, and captured Composer package
          versions.
        </li>
        <li>
          <icon-fa6-solid:code-branch class="icon" />
          Link deployments to commits automatically when your project has a Git Repository URL.
        </li>
      </ul>

      <div class="whats-new-setup">
        <p class="whats-new-setup-title">Quick setup</p>
        <p>
          Create a project API key, set <code>SHOPMON_SHOP_ID</code> and
          <code>SHOPMON_API_KEY</code>, then wrap your deployment command with
          <code>shopmon-cli deploy -- your-command</code>.
        </p>
      </div>
    </template>

    <template #footer>
      <a
        href="https://github.com/FriendsOfShopware/shopmon-cli"
        target="_blank"
        rel="noopener noreferrer"
        class="btn"
        @click="$emit('close')"
      >
        <icon-fa6-solid:arrow-up-right-from-square class="icon" />
        Open shopmon-cli
      </a>

      <router-link :to="{ name: 'account.project.list' }" class="btn" @click="$emit('close')">
        <icon-fa6-solid:folder-open class="icon" />
        Open Projects
      </router-link>

      <button type="button" class="btn btn-primary" @click="$emit('close')">Close</button>
    </template>
  </modal>
</template>

<script setup lang="ts">
import Modal from "@/components/layout/Modal.vue";

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
</style>
