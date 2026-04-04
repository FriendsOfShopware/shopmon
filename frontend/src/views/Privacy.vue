<template>
  <div>
    <header-container :title="$t('privacy.title')" />
    <Card>
      <CardContent>
        <section class="mb-8">
          <h2 class="text-primary mb-4 text-2xl font-semibold">{{ $t("privacy.dataCollect") }}</h2>
          <h3 class="mb-2 mt-4 text-lg font-medium">{{ $t("privacy.accountInfo") }}</h3>
          <ul class="mb-4 pl-6 list-disc">
            <li class="mb-2">{{ $t("privacy.accountEmail") }}</li>
            <li class="mb-2">{{ $t("privacy.accountGithub") }}</li>
            <li class="mb-2">{{ $t("privacy.accountPassword") }}</li>
            <li class="mb-2">{{ $t("privacy.accountPreferences") }}</li>
          </ul>

          <h3 class="mb-2 mt-4 text-lg font-medium">{{ $t("privacy.shopMonitoring") }}</h3>
          <ul class="mb-4 pl-6 list-disc">
            <li class="mb-2">{{ $t("privacy.shopUrls") }}</li>
            <li class="mb-2">{{ $t("privacy.shopExtensions") }}</li>
            <li class="mb-2">{{ $t("privacy.shopTasks") }}</li>
          </ul>

          <h3 class="mb-2 mt-4 text-lg font-medium">{{ $t("privacy.technicalData") }}</h3>
          <ul class="mb-4 pl-6 list-disc">
            <li class="mb-2">{{ $t("privacy.techLogs") }}</li>
            <li class="mb-2">{{ $t("privacy.techSessions") }}</li>
            <li class="mb-2">{{ $t("privacy.techSentry") }}</li>
          </ul>
        </section>

        <section class="mb-8">
          <h2 class="text-primary mb-4 text-2xl font-semibold">{{ $t("privacy.howWeUse") }}</h2>
          <ul class="mb-4 pl-6 list-disc">
            <li class="mb-2">{{ $t("privacy.useMonitoring") }}</li>
            <li class="mb-2">{{ $t("privacy.useNotifications") }}</li>
            <li class="mb-2">{{ $t("privacy.useImprove") }}</li>
            <li class="mb-2">{{ $t("privacy.useSecurity") }}</li>
          </ul>
        </section>

        <section class="mb-8">
          <h2 class="text-primary mb-4 text-2xl font-semibold">{{ $t("privacy.storage") }}</h2>
          <ul class="mb-4 pl-6 list-disc">
            <li class="mb-2">{{ $t("privacy.storageServer") }}</li>
            <li class="mb-2">{{ $t("privacy.storageEncryption") }}</li>
            <li class="mb-2">{{ $t("privacy.storageBackups") }}</li>
            <li class="mb-2">{{ $t("privacy.storageSeparation") }}</li>
            <li class="mb-2">{{ $t("privacy.storageDeletion") }}</li>
          </ul>
        </section>

        <section class="mb-8">
          <h2 class="text-primary mb-4 text-2xl font-semibold">{{ $t("privacy.cookies") }}</h2>
          <p class="mb-4 text-muted-foreground">{{ $t("privacy.cookiesDesc") }}</p>
        </section>

        <section>
          <h2 class="text-primary mb-4 text-2xl font-semibold">{{ $t("privacy.contact") }}</h2>
          <p class="text-muted-foreground">{{ $t("privacy.contactDesc") }}</p>
        </section>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts"></script>
