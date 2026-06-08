<template>
  <div class="space-y-6">
    <!-- Header -->
    <PageHeader :title="$t('shop.listTitle')">
      <Button size="sm" as-child>
        <RouterLink :to="{ name: 'account.shops.new' }">
          <icon-fa6-solid:folder-plus class="mr-1.5 size-3" />
          {{ $t("shop.addShop") }}
        </RouterLink>
      </Button>
    </PageHeader>

    <template v-if="!loading">
      <!-- Empty state -->
      <EmptyState
        v-if="shops.length === 0"
        :icon="IconFolder"
        :title="$t('shop.noShops')"
        :description="$t('shop.getStarted')"
      >
        <Button as-child>
          <RouterLink :to="{ name: 'account.shops.new' }">
            <icon-fa6-solid:plus class="mr-1.5 size-3" />
            {{ $t("shop.addShop") }}
          </RouterLink>
        </Button>
      </EmptyState>

      <!-- Shop list -->
      <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <RouterLink
          v-for="shop in shops"
          :key="shop.id"
          :to="shopLink(shop)"
          class="group relative flex items-start gap-3 rounded-xl border bg-card p-4 shadow-sm transition-all duration-200 hover:border-primary/30 hover:shadow-md"
        >
          <div class="flex size-10 shrink-0 items-center justify-center rounded-lg border bg-muted">
            <img
              v-if="defaultEnv(shop)?.favicon"
              :src="defaultEnv(shop)!.favicon!"
              alt=""
              class="size-5 rounded"
            />
            <icon-fa6-solid:folder v-else class="size-4 text-muted-foreground/50" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span
                class="truncate font-semibold leading-tight transition-colors group-hover:text-primary"
              >
                {{ shop.name }}
              </span>
              <StatusIcon v-if="defaultEnv(shop)" :status="defaultEnv(shop)!.status" />
            </div>
            <p v-if="shop.description" class="mt-0.5 truncate text-xs text-muted-foreground">
              {{ shop.description }}
            </p>
            <div class="mt-1.5 flex items-center gap-2">
              <Badge v-if="defaultEnv(shop)" variant="secondary" class="font-mono text-xs">
                {{ defaultEnv(shop)!.shopwareVersion }}
              </Badge>
              <span v-if="envCount(shop) > 1" class="text-xs text-muted-foreground">
                {{ $t("shop.envCount", { count: envCount(shop) }) }}
              </span>
            </div>
          </div>

          <!-- Edit button -->
          <Button
            variant="ghost"
            size="icon"
            class="absolute top-2 right-2 size-7 opacity-0 transition-opacity group-hover:opacity-100"
            as="span"
            :title="$t('nav.editShop')"
            @click.prevent="
              $router.push({ name: 'account.shops.edit', params: { shopId: shop.id } })
            "
          >
            <icon-fa6-solid:pencil class="size-3" />
          </Button>
        </RouterLink>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import StatusIcon from "@/components/StatusIcon.vue";
import PageHeader from "@/components/PageHeader.vue";
import EmptyState from "@/components/EmptyState.vue";
import IconFolder from "~icons/fa6-solid/folder";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { ref } from "vue";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";

type AccountShop = components["schemas"]["AccountShop"];
type AccountEnvironment = components["schemas"]["AccountEnvironment"];

const loading = ref(true);
const { environments } = useAccountEnvironments();
const shops = ref<AccountShop[]>([]);

function defaultEnv(shop: AccountShop): AccountEnvironment | undefined {
  return environments.value.find((e) => e.id === shop.defaultEnvironmentId);
}

function envCount(shop: AccountShop): number {
  return environments.value.filter((e) => e.shopId === shop.id).length;
}

function shopLink(shop: AccountShop) {
  return {
    name: "account.environments.detail",
    params: { environmentId: shop.defaultEnvironmentId },
  };
}

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
