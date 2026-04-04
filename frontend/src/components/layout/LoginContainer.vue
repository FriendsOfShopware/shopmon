<template>
  <div class="login-container">
    <div class="login-toggles">
      <button class="dark-mode-toggle" type="button" @click="toggleDarkMode">
        <icon-fa6-regular:moon v-if="darkMode" class="icon" />
        <icon-octicon:sun-16 v-else class="icon" />
      </button>

      <button class="locale-toggle" type="button" @click="toggleLocale">
        {{ String(locale) === "en" ? "DE" : "EN" }}
      </button>
    </div>

    <div class="login-content">
      <router-link :to="{ name: 'home' }">
        <logo class="logo" />
      </router-link>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useDarkMode } from "@/composables/useDarkMode";
import { useLocale } from "@/composables/useLocale";

const { darkMode, toggleDarkMode } = useDarkMode();
const { locale, toggleLocale } = useLocale();
</script>

<style>
.login-container {
  position: relative;
  min-height: calc(100vh - 82px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;

  .btn {
    .icon {
      width: 1.25rem;
      height: 1.25rem;
    }
  }
}

.login-toggles {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.dark-mode-toggle,
.locale-toggle {
  height: 2rem;
  width: 2rem;
  opacity: 0.8;
  transition: opacity 0.2s ease-in-out;
  background: none;
  border: none;
  cursor: pointer;
  outline: none;

  &:hover {
    opacity: 1;
  }

  .icon {
    width: 1.25rem;
    height: 1.25rem;
  }
}

.locale-toggle {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.05em;
}

.login-content {
  max-width: 28rem;
  width: 100%;
  min-height: 37rem;
}

.logo {
  height: 9rem;
  width: auto;
  margin: 0 auto;
}

.login-header {
  text-align: center;
  margin: 2rem 0;

  h2 {
    font-size: 1.875rem;
    font-weight: bold;
    line-height: 1.25;
    margin-bottom: 0.5rem;
  }

  p {
    color: var(--text-color-muted);
  }
}

.login-form-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  text-align: center;
}

.login-form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
</style>
