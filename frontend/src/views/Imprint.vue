<template>
  <!-- Page header -->
  <div class="bg-gradient-to-b from-[#0a3d5c] via-primary to-primary/30 pb-12 pt-28 sm:pt-32">
    <div class="mx-auto max-w-3xl px-6 text-center">
      <h1 class="text-3xl font-bold text-white">{{ $t("imprint.title") }}</h1>
    </div>
  </div>

  <!-- Content -->
  <div class="mx-auto -mt-6 max-w-3xl px-6 pb-16">
    <Card>
      <CardContent class="py-8">
        <h2 class="mb-3 text-xl font-semibold text-primary">{{ $t("imprint.contact") }}</h2>
        <p class="mb-4 text-muted-foreground">{{ $t("imprint.description") }}</p>
        <p class="text-muted-foreground">{{ $t("imprint.email") }}</p>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { Card, CardContent } from "@/components/ui/card";
</script>
