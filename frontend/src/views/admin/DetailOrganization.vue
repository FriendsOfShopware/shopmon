<template>
  <div class="space-y-6">
    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-12 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ t("admin.loadingOrganization") }}
    </div>

    <!-- Error -->
    <Alert v-else-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Not found -->
    <EmptyState
      v-else-if="!organization"
      :icon="IconBuilding"
      :title="t('admin.organizationNotFound')"
    />

    <!-- Detail -->
    <template v-else>
      <PageHeader :title="organization.name">
        <Button variant="outline" size="sm" @click="router.back()">
          <icon-fa6-solid:arrow-left class="mr-1.5 size-3" />
          {{ t("admin.back") }}
        </Button>
      </PageHeader>

      <p class="-mt-4 text-sm text-muted-foreground">{{ organization.slug }}</p>

      <!-- Stats -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard :icon="IconUsers" :value="organization.memberCount" :label="t('admin.members')" />
        <StatCard
          :icon="IconEarthAmericas"
          :value="organization.environmentCount"
          :label="t('common.environments')"
        />
        <StatCard
          :icon="IconEnvelope"
          :value="organization.invitations.length"
          :label="t('admin.invitations')"
        />
        <StatCard
          :icon="IconCalendar"
          :value="formatDate(organization.createdAt)"
          :label="t('admin.created')"
        />
      </div>

      <!-- Bento grid -->
      <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <!-- Members -->
        <CardSection :icon="IconUsers" :title="t('admin.members')">
          <EmptyState
            v-if="organization.members.length === 0"
            :icon="IconUsers"
            :title="t('admin.noMembers')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <RouterLink
              v-for="member in organization.members"
              :key="member.userId"
              :to="{ name: 'admin.users.detail', params: { id: member.userId } }"
              class="group flex items-center gap-3 rounded-xl border bg-card px-4 py-3 transition-all duration-200 hover:border-primary/30 hover:shadow-sm"
            >
              <div class="min-w-0 flex-1">
                <div class="truncate font-medium transition-colors group-hover:text-primary">
                  {{ member.name }}
                </div>
                <div class="truncate text-xs text-muted-foreground">{{ member.email }}</div>
              </div>
              <Badge variant="secondary" class="shrink-0 text-xs capitalize">{{
                member.role
              }}</Badge>
            </RouterLink>
          </div>
        </CardSection>

        <!-- Shops & Environments -->
        <CardSection
          :icon="IconStore"
          :title="t('admin.shopsAndEnvironments')"
          class="lg:col-span-2"
        >
          <EmptyState
            v-if="shopGroups.length === 0"
            :icon="IconStore"
            :title="t('admin.noShops')"
            size="sm"
          />
          <div v-else class="space-y-4">
            <div v-for="shop in shopGroups" :key="shop.id" class="rounded-xl border bg-card">
              <!-- Shop header -->
              <div class="flex items-center gap-3 border-b px-4 py-3">
                <div class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted">
                  <IconStore class="size-3.5 text-muted-foreground" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="truncate font-medium">{{ shop.name }}</div>
                  <div v-if="shop.description" class="truncate text-xs text-muted-foreground">
                    {{ shop.description }}
                  </div>
                </div>
                <Badge variant="secondary" class="shrink-0 text-xs">
                  {{ t("admin.nEnvironments", { count: shop.environments.length }) }}
                </Badge>
              </div>

              <!-- Shop's environments -->
              <p
                v-if="shop.environments.length === 0"
                class="px-4 py-3 text-xs text-muted-foreground"
              >
                {{ t("admin.noEnvironments") }}
              </p>
              <div v-else class="divide-y">
                <RouterLink
                  v-for="env in shop.environments"
                  :key="env.id"
                  :to="{ name: 'admin.environments.detail', params: { id: env.id } }"
                  class="group flex items-center gap-3 px-4 py-2.5 transition-colors hover:bg-accent/50"
                >
                  <StatusIcon :status="env.status" />
                  <span
                    class="min-w-0 flex-1 truncate text-sm transition-colors group-hover:text-primary"
                  >
                    {{ env.name }}
                  </span>
                  <Badge variant="secondary" class="shrink-0 font-mono text-[10px]">
                    {{ env.shopwareVersion }}
                  </Badge>
                </RouterLink>
              </div>
            </div>
          </div>
        </CardSection>

        <!-- Invitations -->
        <CardSection :icon="IconEnvelope" :title="t('admin.invitations')">
          <EmptyState
            v-if="organization.invitations.length === 0"
            :icon="IconEnvelope"
            :title="t('admin.noInvitations')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <div
              v-for="inv in organization.invitations"
              :key="inv.id"
              class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3"
            >
              <div class="min-w-0 flex-1">
                <div class="truncate font-medium">{{ inv.email }}</div>
                <div class="truncate text-xs text-muted-foreground">
                  {{ t("admin.invitedBy", { name: inv.inviterName }) }}
                </div>
              </div>
              <div class="flex shrink-0 flex-col items-end gap-1">
                <Badge variant="secondary" class="text-xs capitalize">{{ inv.status }}</Badge>
                <span class="text-xs tabular-nums text-muted-foreground">
                  {{ formatDate(inv.expiresAt) }}
                </span>
              </div>
            </div>
          </div>
        </CardSection>

        <!-- SSO Providers -->
        <CardSection :icon="IconShieldHalved" :title="t('admin.ssoProviders')">
          <EmptyState
            v-if="organization.ssoProviders.length === 0"
            :icon="IconShieldHalved"
            :title="t('admin.noSSOProviders')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <div
              v-for="provider in organization.ssoProviders"
              :key="provider.id"
              class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3"
            >
              <div class="min-w-0 flex-1">
                <div class="truncate font-medium">{{ provider.domain }}</div>
                <div class="truncate text-xs text-muted-foreground">{{ provider.issuer }}</div>
              </div>
            </div>
          </div>
        </CardSection>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import { api } from "@/helpers/api";
import { formatDate } from "@/helpers/formatter";
import type { components } from "@/types/api";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import PageHeader from "@/components/PageHeader.vue";
import StatCard from "@/components/StatCard.vue";
import CardSection from "@/components/CardSection.vue";
import EmptyState from "@/components/EmptyState.vue";
import StatusIcon from "@/components/StatusIcon.vue";

import IconUsers from "~icons/fa6-solid/users";
import IconEarthAmericas from "~icons/fa6-solid/earth-americas";
import IconEnvelope from "~icons/fa6-solid/envelope";
import IconCalendar from "~icons/fa6-solid/calendar";
import IconShieldHalved from "~icons/fa6-solid/shield-halved";
import IconStore from "~icons/fa6-solid/store";
import IconBuilding from "~icons/fa6-solid/building";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

type OrgDetail = components["schemas"]["AdminOrganizationDetail"];
type OrgEnvironment = OrgDetail["environments"][number];

const organization = ref<OrgDetail | null>(null);
const loading = ref(true);
const error = ref("");

// Group environments under their shop, preserving the shop ordering from the
// API (shops without environments still appear).
const shopGroups = computed(() => {
  const org = organization.value;
  if (!org) return [];

  const envsByShop = new Map<number, OrgEnvironment[]>();
  for (const env of org.environments) {
    const list = envsByShop.get(env.shopId) ?? [];
    list.push(env);
    envsByShop.set(env.shopId, list);
  }

  return org.shops.map((shop) => ({
    id: shop.id,
    name: shop.name,
    description: shop.description,
    environments: envsByShop.get(shop.id) ?? [],
  }));
});

async function loadOrganization() {
  loading.value = true;
  error.value = "";

  const { data, error: respError } = await api.GET("/admin/organizations/{orgId}", {
    params: { path: { orgId: String(route.params.id) } },
  });

  if (respError) {
    error.value = t("admin.failedLoadOrgs", {
      error: (respError as { message?: string }).message ?? "",
    });
  } else if (data) {
    organization.value = data;
  }

  loading.value = false;
}

onMounted(loadOrganization);
</script>
