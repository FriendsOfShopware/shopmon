<template>
  <div class="space-y-6">
    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-12 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ t("admin.loadingUser") }}
    </div>

    <!-- Not found -->
    <EmptyState v-else-if="!user" :icon="IconUserSlash" :title="t('admin.userNotFound')" />

    <!-- Content -->
    <template v-else>
      <PageHeader :title="user.name">
        <Button variant="outline" size="sm" @click="router.back()">
          <icon-fa6-solid:arrow-left class="mr-1.5 size-3" />
          {{ t("admin.back") }}
        </Button>
      </PageHeader>

      <Alert v-if="error" variant="destructive">
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <!-- Stat cards -->
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          :icon="IconUserShield"
          :value="user.role"
          :label="t('common.role')"
          :color="user.role === 'admin' ? 'primary' : 'muted'"
        />
        <StatCard
          :icon="IconCircleCheck"
          :value="statusLabel"
          :label="t('admin.status')"
          :color="user.banned ? 'destructive' : !user.emailVerified ? 'warning' : 'success'"
        />
        <StatCard
          :icon="IconBuilding"
          :value="user.memberships.length"
          :label="t('admin.memberships')"
        />
        <StatCard
          :icon="IconCalendar"
          :value="formatDate(user.createdAt)"
          :label="t('admin.memberSince')"
        />
      </div>

      <!-- Bento grid -->
      <div class="grid gap-4 lg:grid-cols-2">
        <!-- Memberships -->
        <CardSection :icon="IconBuilding" :title="t('admin.memberships')">
          <EmptyState
            v-if="user.memberships.length === 0"
            :icon="IconBuilding"
            :title="t('admin.noMemberships')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <RouterLink
              v-for="m in user.memberships"
              :key="m.organizationId"
              :to="{ name: 'admin.organizations.detail', params: { id: m.organizationId } }"
              class="group flex items-center gap-3 rounded-xl border bg-card px-4 py-3 transition-colors hover:border-primary/30"
            >
              <div
                class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10"
              >
                <icon-fa6-solid:building class="size-3.5 text-primary" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="truncate font-medium transition-colors group-hover:text-primary">
                  {{ m.organizationName }}
                </div>
                <div class="text-xs text-muted-foreground tabular-nums">
                  {{ formatDate(m.createdAt) }}
                </div>
              </div>
              <Badge variant="secondary" class="shrink-0 text-xs capitalize">{{ m.role }}</Badge>
            </RouterLink>
          </div>
        </CardSection>

        <!-- Sessions -->
        <CardSection :icon="IconDesktop" :title="t('admin.sessions')">
          <EmptyState
            v-if="user.sessions.length === 0"
            :icon="IconDesktop"
            :title="t('admin.noSessions')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <div
              v-for="s in user.sessions"
              :key="s.id"
              class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3"
            >
              <div class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted">
                <icon-fa6-solid:desktop class="size-3.5 text-muted-foreground" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <span class="font-medium tabular-nums">{{ s.ipAddress || "—" }}</span>
                  <Badge
                    v-if="s.impersonated"
                    class="bg-warning/10 text-warning border-warning/20 text-[10px]"
                  >
                    {{ t("admin.impersonatedSession") }}
                  </Badge>
                </div>
                <div
                  class="mt-0.5 truncate text-xs text-muted-foreground"
                  :title="s.userAgent ?? ''"
                >
                  {{ s.userAgent || "—" }}
                </div>
              </div>
              <div class="shrink-0 text-xs text-muted-foreground tabular-nums">
                {{ formatDateTime(s.createdAt) }}
              </div>
            </div>
          </div>
        </CardSection>

        <!-- Auth providers -->
        <CardSection :icon="IconKey" :title="t('admin.authProviders')">
          <EmptyState
            v-if="user.authProviders.length === 0"
            :icon="IconKey"
            :title="t('admin.noAuthProviders')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <div
              v-for="p in user.authProviders"
              :key="p.id"
              class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3"
            >
              <div
                class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10"
              >
                <icon-fa6-solid:key class="size-3.5 text-primary" />
              </div>
              <div class="min-w-0 flex-1">
                <Badge variant="secondary" class="text-xs capitalize">{{ p.providerId }}</Badge>
              </div>
              <div class="shrink-0 text-xs text-muted-foreground tabular-nums">
                {{ formatDate(p.createdAt) }}
              </div>
            </div>
          </div>
        </CardSection>

        <!-- Audit trail -->
        <CardSection :icon="IconClock" :title="t('admin.auditLog')">
          <EmptyState
            v-if="user.auditLog.length === 0"
            :icon="IconClock"
            :title="t('admin.noAuditEntries')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <div
              v-for="entry in user.auditLog"
              :key="entry.id"
              class="rounded-xl border bg-card px-4 py-3"
            >
              <div class="flex items-center justify-between gap-2">
                <span class="font-medium">{{ actionLabel(entry.action) }}</span>
                <span class="shrink-0 text-xs text-muted-foreground tabular-nums">
                  {{ formatDateTime(entry.createdAt) }}
                </span>
              </div>
              <div
                v-if="entry.actorName || entry.actorEmail"
                class="mt-0.5 text-xs text-muted-foreground"
              >
                {{ entry.actorName || entry.actorEmail }}
              </div>
              <div v-if="entry.detail" class="mt-1 text-xs text-muted-foreground">
                {{ entry.detail }}
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
import { formatDate, formatDateTime } from "@/helpers/formatter";
import type { components } from "@/types/api";

import PageHeader from "@/components/PageHeader.vue";
import StatCard from "@/components/StatCard.vue";
import CardSection from "@/components/CardSection.vue";
import EmptyState from "@/components/EmptyState.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import IconUserShield from "~icons/fa6-solid/user-shield";
import IconCircleCheck from "~icons/fa6-solid/circle-check";
import IconBuilding from "~icons/fa6-solid/building";
import IconCalendar from "~icons/fa6-solid/calendar";
import IconDesktop from "~icons/fa6-solid/desktop";
import IconKey from "~icons/fa6-solid/key";
import IconClock from "~icons/fa6-solid/clock-rotate-left";
import IconUserSlash from "~icons/fa6-solid/user-slash";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const id = route.params.id as string;

const user = ref<components["schemas"]["AdminUserDetail"] | null>(null);
const loading = ref(true);
const error = ref("");

const statusLabel = computed(() => {
  if (!user.value) return "";
  if (user.value.banned) return t("admin.banned");
  if (!user.value.emailVerified) return t("admin.unverified");
  return t("admin.active");
});

const actionLabels: Record<string, string> = {
  "admin.set_role": "actionSetRole",
  "admin.ban_user": "actionBanUser",
  "admin.unban_user": "actionUnbanUser",
  "admin.impersonate": "actionImpersonate",
  "user.password_change": "actionPasswordChange",
  "user.password_reset": "actionPasswordReset",
};

function actionLabel(action: string): string {
  const key = actionLabels[action];
  return key ? t(`admin.${key}`) : action;
}

async function loadUser() {
  loading.value = true;
  error.value = "";
  try {
    const { data, error: respError } = await api.GET("/auth/admin/users/{userId}", {
      params: { path: { userId: id } },
    });
    if (respError) {
      error.value =
        (respError as unknown as { message?: string })?.message ?? t("admin.userNotFound");
      user.value = null;
      return;
    }
    user.value = data ?? null;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
    user.value = null;
  } finally {
    loading.value = false;
  }
}

onMounted(loadUser);
</script>
