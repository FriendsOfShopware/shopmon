<template>
  <div>
    <header-container :title="$t('imprint.title')" />

    <Card>
      <CardContent>
        <section>
          <h2 class="text-primary mb-4 text-2xl font-semibold">{{ $t("imprint.contact") }}</h2>
          <div>
            <p class="mb-4 text-muted-foreground">{{ $t("imprint.description") }}</p>
            <p class="text-muted-foreground">{{ $t("imprint.email") }}</p>
          </div>
        </section>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts"></script>
