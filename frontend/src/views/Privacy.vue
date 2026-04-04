<template>
  <!-- Page header -->
  <div class="bg-gradient-to-b from-[#0a3d5c] via-primary to-primary/30 pb-12 pt-28 sm:pt-32">
    <div class="mx-auto max-w-3xl px-6 text-center">
      <h1 class="text-3xl font-bold text-white">{{ $t('privacy.title') }}</h1>
    </div>
  </div>

  <!-- Content -->
  <div class="mx-auto -mt-6 max-w-3xl px-6 pb-16">
    <Card>
      <CardContent class="prose prose-sm dark:prose-invert max-w-none py-8 [&_h2]:mb-3 [&_h2]:mt-8 [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:text-primary [&_h2:first-child]:mt-0 [&_h3]:mb-2 [&_h3]:mt-5 [&_h3]:text-base [&_h3]:font-medium [&_ul]:mb-4 [&_ul]:space-y-1.5 [&_li]:text-muted-foreground [&_p]:text-muted-foreground">
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

        <h2>{{ $t("privacy.howWeUse") }}</h2>
        <ul>
          <li>{{ $t("privacy.useMonitoring") }}</li>
          <li>{{ $t("privacy.useNotifications") }}</li>
          <li>{{ $t("privacy.useImprove") }}</li>
          <li>{{ $t("privacy.useSecurity") }}</li>
        </ul>

        <h2>{{ $t("privacy.storage") }}</h2>
        <ul>
          <li>{{ $t("privacy.storageServer") }}</li>
          <li>{{ $t("privacy.storageEncryption") }}</li>
          <li>{{ $t("privacy.storageBackups") }}</li>
          <li>{{ $t("privacy.storageSeparation") }}</li>
          <li>{{ $t("privacy.storageDeletion") }}</li>
        </ul>

        <h2>{{ $t("privacy.cookies") }}</h2>
        <p>{{ $t("privacy.cookiesDesc") }}</p>

        <h2>{{ $t("privacy.contact") }}</h2>
        <p>{{ $t("privacy.contactDesc") }}</p>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { Card, CardContent } from "@/components/ui/card";
</script>
