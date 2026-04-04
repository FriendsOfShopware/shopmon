<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">{{ $t("shop.listTitle") }}</h1>
      <Button size="sm" as-child>
        <RouterLink :to="{ name: 'account.shops.new' }">
          <icon-fa6-solid:folder-plus class="mr-1.5 size-3" />
          {{ $t("shop.addShop") }}
        </RouterLink>
      </Button>
    </div>

    <template v-if="!loading">
      <!-- Empty state -->
      <div
        v-if="shops.length === 0"
        class="flex flex-col items-center gap-4 rounded-xl border border-dashed py-20 text-center"
      >
        <div class="flex size-14 items-center justify-center rounded-2xl bg-primary/10">
          <icon-fa6-solid:folder class="size-6 text-primary" />
        </div>
        <h2 class="text-xl font-semibold">{{ $t("shop.noShops") }}</h2>
        <p class="max-w-md text-muted-foreground">{{ $t("shop.getStarted") }}</p>
        <Button as-child>
          <RouterLink :to="{ name: 'account.shops.new' }">
            <icon-fa6-solid:plus class="mr-1.5 size-3" />
            {{ $t("shop.addShop") }}
          </RouterLink>
        </Button>
      </div>

      <!-- Shop list -->
      <div v-else class="space-y-4">
        <Card v-for="shop in shops" :key="shop.id">
          <CardHeader class="pb-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="flex size-9 items-center justify-center rounded-lg bg-primary/10">
                  <icon-fa6-solid:folder class="size-4 text-primary" />
                </div>
                <div>
                  <CardTitle class="text-base">{{ shop.name }}</CardTitle>
                  <CardDescription v-if="shop.description" class="mt-0.5">{{
                    shop.description
                  }}</CardDescription>
                </div>
              </div>
              <Button variant="ghost" size="sm" as-child>
                <RouterLink :to="{ name: 'account.shops.edit', params: { shopId: shop.id } }">
                  <icon-fa6-solid:pen-to-square class="mr-1.5 size-3" />
                  {{ $t("nav.editShop") }}
                </RouterLink>
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <!-- Environments grid -->
            <div
              v-if="shopEnvironments[shop.id]?.length > 0"
              class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3"
            >
              <RouterLink
                v-for="env in shopEnvironments[shop.id]"
                :key="env.id"
                :to="{
                  name: 'account.environments.detail',
                  params: { organizationId: env.organizationId, environmentId: env.id },
                }"
                class="group flex items-center gap-3 rounded-xl border p-3 transition-all duration-200 hover:border-primary/30 hover:shadow-sm"
              >
                <div
                  class="flex size-9 shrink-0 items-center justify-center rounded-lg border bg-muted"
                >
                  <img v-if="env.favicon" :src="env.favicon" alt="" class="size-5 rounded" />
                  <icon-fa6-solid:earth-americas v-else class="size-3.5 text-muted-foreground/50" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span
                      class="truncate text-sm font-medium group-hover:text-primary transition-colors"
                      >{{ env.name }}</span
                    >
                    <StatusIcon :status="env.status" />
                  </div>
                  <div class="mt-0.5 flex items-center gap-2 text-xs text-muted-foreground">
                    <Badge variant="secondary" class="font-mono text-[10px]">{{
                      env.shopwareVersion
                    }}</Badge>
                  </div>
                </div>
                <icon-fa6-solid:chevron-right
                  class="size-3 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
                />
              </RouterLink>

              <!-- Add environment card -->
              <RouterLink
                :to="{ name: 'account.environments.new', query: { shopId: shop.id } }"
                class="flex items-center gap-3 rounded-xl border border-dashed p-3 text-muted-foreground transition-colors hover:border-primary/30 hover:bg-accent hover:text-foreground"
              >
                <div class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted">
                  <icon-fa6-solid:plus class="size-3.5" />
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-medium">{{ $t("environment.addEnvironment") }}</div>
                </div>
              </RouterLink>
            </div>

            <!-- No environments yet -->
            <div
              v-else
              class="flex flex-col items-center gap-3 rounded-xl border border-dashed bg-muted/30 py-8 text-center"
            >
              <p class="text-sm text-muted-foreground">{{ $t("shop.noEnvironmentsYet") }}</p>
              <Button size="sm" as-child>
                <RouterLink :to="{ name: 'account.environments.new', query: { shopId: shop.id } }">
                  <icon-fa6-solid:plus class="mr-1.5 size-3" />
                  {{ $t("environment.addEnvironment") }}
                </RouterLink>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import StatusIcon from "@/components/StatusIcon.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { computed, ref } from "vue";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";

const loading = ref(true);
const { environments } = useAccountEnvironments();
const shops = ref<components["schemas"]["AccountShop"][]>([]);

const shopEnvironments = computed(() => {
  const grouped: Record<number, typeof environments.value> = {};
  for (const env of environments.value) {
    const shopId = env.shopId;
    if (shopId) {
      grouped[shopId] ??= [];
      grouped[shopId].push(env);
    }
  }
  return grouped;
});

api
  .GET("/account/shops")
  .then((shopsRes) => {
    if (shopsRes.data) shops.value = shopsRes.data;
    loading.value = false;
  })
  .catch(() => {
    loading.value = false;
  });
</script>
