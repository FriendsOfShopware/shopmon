<template>
  <header class="flex items-start justify-between mb-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{{ $t('shop.listTitle') }}</h1>
    </div>
    <div class="flex gap-2 items-start">
      <Button as-child>
        <RouterLink :to="{ name: 'account.shops.new' }">
          <icon-fa6-solid:folder-plus class="size-4" aria-hidden="true" />
          {{ $t("shop.addShop") }}
        </RouterLink>
      </Button>
    </div>
  </header>

  <div v-if="!loading" class="space-y-6">
    <Card v-if="shops.length === 0">
      <CardContent class="py-16 flex flex-col items-center gap-6 text-center">
        <svg
          class="size-12 text-muted-foreground"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            vector-effect="non-scaling-stroke"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
          />
        </svg>
        <h2 class="text-2xl font-semibold">{{ $t('shop.noShops') }}</h2>
        <p class="text-muted-foreground max-w-lg">{{ $t("shop.getStarted") }}</p>
        <Button as-child>
          <RouterLink :to="{ name: 'account.shops.new' }">
            <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
            {{ $t("shop.addShop") }}
          </RouterLink>
        </Button>
      </CardContent>
    </Card>

    <div v-else class="space-y-6">
      <!-- Shops -->
      <Card
        v-for="shop in shops"
        :key="shop.id"
      >
        <CardHeader>
          <CardTitle>{{ shop.name }}</CardTitle>
          <CardDescription v-if="shop.description">{{ shop.description }}</CardDescription>
          <CardAction>
            <Button variant="outline" as-child>
              <RouterLink
                :to="{ name: 'account.shops.edit', params: { shopId: shop.id } }"
              >
                <icon-fa6-solid:pen-to-square class="size-4" aria-hidden="true" />
                {{ $t("nav.editShop") }}
              </RouterLink>
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <p class="flex items-center gap-2 text-sm text-muted-foreground mb-4">
            <span>{{
              $t("shop.nEnvironments", { count: shopEnvironments[shop.id]?.length || 0 })
            }}</span>
          </p>

          <div v-if="shopEnvironments[shop.id]?.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <div v-for="env in shopEnvironments[shop.id]" :key="env.id" class="relative group">
              <router-link
                :to="{
                  name: 'account.environments.detail',
                  params: {
                    organizationId: env.organizationId,
                    environmentId: env.id,
                  },
                }"
                class="flex items-center gap-3 p-3 rounded-lg border hover:bg-accent transition-colors"
              >
                <div class="shrink-0 size-10 flex items-center justify-center rounded-md bg-muted">
                  <img
                    v-if="env.favicon"
                    :src="env.favicon"
                    alt="Environment favicon"
                    class="size-6 rounded"
                  />
                  <icon-fa6-solid:store v-else class="size-4 text-muted-foreground" />
                </div>

                <div class="flex-1 min-w-0 pr-7">
                  <div class="font-medium truncate">
                    {{ env.name }}
                  </div>
                  <div class="text-sm text-muted-foreground">{{ env.shopwareVersion }}</div>
                  <status-icon :status="env.status" class="mt-1" />
                </div>
              </router-link>

              <div class="absolute top-3 right-3">
                <a
                  :href="env.url"
                  target="_blank"
                  :title="$t('shop.visitEnvironment')"
                  class="text-muted-foreground hover:text-foreground"
                >
                  <icon-fa6-solid:arrow-up-right-from-square class="size-4" />
                </a>
              </div>
            </div>

            <router-link
              :to="{ name: 'account.environments.new', query: { shopId: shop.id } }"
              class="flex items-center gap-3 p-3 rounded-lg border border-dashed hover:bg-accent transition-colors min-h-[70px]"
            >
              <div class="shrink-0 size-10 flex items-center justify-center rounded-md bg-muted text-muted-foreground">
                <icon-fa6-solid:plus class="size-4" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="font-medium">{{ $t("environment.addEnvironment") }}</div>
                <div class="text-sm text-muted-foreground">{{ $t("shop.addToShop") }}</div>
              </div>
            </router-link>
          </div>

          <div v-else class="rounded-xl border border-dashed bg-muted/50 py-10 flex flex-col items-center gap-4 text-center">
            <p class="text-muted-foreground">{{ $t("shop.noEnvironmentsYet") }}</p>
            <Button as-child>
              <RouterLink :to="{ name: 'account.environments.new', query: { shopId: shop.id } }">
                <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
                {{ $t("environment.addEnvironment") }}
              </RouterLink>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { computed, ref } from "vue";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";

const loading = ref(true);
const { environments } = useAccountEnvironments();
const shops = ref<components["schemas"]["AccountShop"][]>([]);

// Group environments by shop
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

// Load shops
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
