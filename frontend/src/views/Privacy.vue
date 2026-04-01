<template>
  <div class="container">
    <header-container :title="$t('privacy.title')" />
    <Panel>
      <section>
        <h2>{{ $t("privacy.dataCollect") }}</h2>
        <h3>{{ $t("privacy.accountInfo") }}</h3>
        <ul>
          <li>{{ $t("privacy.accountEmail") }}</li>
          <li>{{ $t("privacy.accountGithub") }}</li>
          <li>{{ $t("privacy.accountPassword") }}</li>
          <li>{{ $t("privacy.accountPreferences") }}</li>
        </ul>

        <h3>{{ $t("privacy.shopMonitoring") }}</h3>
        <ul>
          <li>{{ $t("privacy.shopUrls") }}</li>
          <li>{{ $t("privacy.shopExtensions") }}</li>
          <li>{{ $t("privacy.shopTasks") }}</li>
        </ul>

        <h3>{{ $t("privacy.technicalData") }}</h3>
        <ul>
          <li>{{ $t("privacy.techLogs") }}</li>
          <li>{{ $t("privacy.techSessions") }}</li>
          <li>{{ $t("privacy.techSentry") }}</li>
        </ul>
      </section>

      <section>
        <h2>{{ $t("privacy.howWeUse") }}</h2>
        <ul>
          <li>{{ $t("privacy.useMonitoring") }}</li>
          <li>{{ $t("privacy.useNotifications") }}</li>
          <li>{{ $t("privacy.useImprove") }}</li>
          <li>{{ $t("privacy.useSecurity") }}</li>
        </ul>
      </section>

      <section>
        <h2>{{ $t("privacy.storage") }}</h2>
        <ul>
          <li>{{ $t("privacy.storageServer") }}</li>
          <li>{{ $t("privacy.storageEncryption") }}</li>
          <li>{{ $t("privacy.storageBackups") }}</li>
          <li>{{ $t("privacy.storageSeparation") }}</li>
          <li>{{ $t("privacy.storageDeletion") }}</li>
        </ul>
      </section>

      <section>
        <h2>{{ $t("privacy.cookies") }}</h2>
        <p>{{ $t("privacy.cookiesDesc") }}</p>
      </section>

      <section>
        <h2>{{ $t("privacy.contact") }}</h2>
        <p>{{ $t("privacy.contactDesc") }}</p>
      </section>
    </Panel>
  </div>
</template>

<script setup lang="ts"></script>
<style scoped>
section {
  margin-bottom: 2rem;
}

h2 {
  color: var(--primary-color);
  margin-bottom: 1rem;
  font-size: 1.5rem;
  font-weight: 600;
}

h3 {
  margin-bottom: 0.5rem;
  margin-top: 1rem;
  font-size: 1.2rem;
  font-weight: 500;
}

ul {
  margin-bottom: 1rem;
  padding-left: 1.5rem;
}

li {
  margin-bottom: 0.5rem;
}

p {
  margin-bottom: 1rem;
  color: var(--text-color);
}
</style>
