<template>
  <div class="app-layout">
    <div class="background-mask" />
    <div class="top-bar">
      <div class="top-bar-container container">
        <div class="top-bar-logo">
          <router-link :to="{ name: 'home' }">
            <logo class="nav-logo-img" />
          </router-link>
        </div>

        <div class="top-bar-actions">
          <router-link
            v-if="session.data"
            :to="{ name: 'account.dashboard' }"
            class="btn btn-primary btn-sm"
          >
            <icon-ri:dashboard-fill class="icon" />
            Dashboard
          </router-link>
          <template v-else>
            <router-link :to="{ name: 'account.login' }" class="btn btn-primary btn-sm">
              <icon-fa6-solid:right-to-bracket class="icon" />
              Login
            </router-link>
            <router-link :to="{ name: 'account.register' }" class="btn btn-primary btn-sm">
              <icon-fa6-solid:user-plus class="icon" />
              Register
            </router-link>
          </template>

          <button class="action action-dark-mode" type="button" @click="toggleDarkMode">
            <icon-fa6-regular:moon v-if="darkMode" class="icon" />

            <icon-octicon:sun-16 v-else class="icon" />
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
import { authClient } from "@/helpers/auth-client";

const session = authClient.useSession();
const { darkMode, toggleDarkMode } = useDarkMode();
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
}
</style>
