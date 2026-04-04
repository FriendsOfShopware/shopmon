<template>
  <div class="app-layout">
    <div class="background-mask" />
    <div class="top-bar">
      <div class="top-bar-container container">
        <div class="top-bar-logo">
          <UiButton :to="{ name: 'home' }">
            <logo class="nav-logo-img" />
          </UiButton>
        </div>

        <div class="top-bar-actions">
          <UiButton
            v-if="session"
            :to="{ name: 'account.dashboard' }"
            variant="primary"
            size="sm"
          >
            <icon-ri:dashboard-fill class="icon" />
            {{ $t("nav.dashboard") }}
          </UiButton>
          <template v-else>
            <UiButton :to="{ name: 'account.login' }" variant="primary" size="sm">
              <icon-fa6-solid:right-to-bracket class="icon" />
              {{ $t("nav.login") }}
            </UiButton>
            <UiButton :to="{ name: 'account.register' }" variant="primary" size="sm">
              <icon-fa6-solid:user-plus class="icon" />
              {{ $t("nav.register") }}
            </UiButton>
          </template>

          <button class="action action-dark-mode" type="button" @click="toggleDarkMode">
            <icon-fa6-regular:moon v-if="darkMode" class="icon" />

            <icon-octicon:sun-16 v-else class="icon" />
          </button>

          <button class="action action-locale" type="button" @click="toggleLocale">
            {{ String(locale) === "en" ? "DE" : "EN" }}
          </button>
        </div>
      </div>
    </div>

    <router-view />

    <layout-footer />
  </div>
</template>

<script setup lang="ts">
import { useDarkMode } from "@/composables/useDarkMode";
import { useSession } from "@/composables/useSession";
import { useLocale } from "@/composables/useLocale";

const { session } = useSession();
const { darkMode, toggleDarkMode } = useDarkMode();
const { locale, toggleLocale } = useLocale();
</script>

<style scoped>
.app-layout {
  min-height: calc(100vh + 1px);
  display: flex;
  flex-direction: column;
}

.background-mask {
  &:after {
    height: 500px;
  }
}

.top-bar-actions {
  margin-left: auto;
  gap: 0;

  .icon {
    margin-left: 0.5rem;
  }

  .action-locale {
    font-size: 0.75rem;
    font-weight: 700;
    letter-spacing: 0.05em;
    color: #bae6fd;
    margin-left: 0.5rem;

    &:hover {
      color: #ffffff;
    }
  }
}
</style>
