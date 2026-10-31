<template>
  <div class="space-y-6">
    <div v-if="loading" class="py-16 text-center text-sm text-muted-foreground">
      {{ $t("common.loading") }}
    </div>

    <div
      v-else-if="loadFailed"
      class="rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3"
    >
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="text-sm font-medium">{{ $t("accounts.loadFailed") }}</div>
        <Button variant="outline" size="sm" @click="load">{{ $t("common.retry") }}</Button>
      </div>
    </div>

    <template v-else-if="data">
      <!-- Summary bar, mirroring the checks tab -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          :icon="IconCircleXmark"
          :value="counts.red"
          :label="$t('accounts.criticalFindings')"
          color="destructive"
        />
        <StatCard
          :icon="IconCircleInfo"
          :value="counts.yellow"
          :label="$t('accounts.warningFindings')"
          color="warning"
        />
        <StatCard
          :icon="IconLightbulb"
          :value="data.postureHints.length"
          :label="$t('accounts.postureTitle')"
          color="muted"
        />
        <StatCard
          :icon="IconUserShield"
          :value="data.accountCount"
          :label="$t('accounts.monitored')"
          color="primary"
        />
      </div>

      <!-- State banners -->
      <div
        v-if="!data.enabled"
        class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-dashed px-4 py-3"
      >
        <div class="text-sm">
          <span class="font-medium">{{ $t("accounts.disabled") }}</span>
          <span class="text-muted-foreground"> — {{ $t("accounts.disabledHint") }}</span>
        </div>
        <Button as-child variant="outline" size="sm">
          <RouterLink :to="settingsRoute">{{ $t("accounts.toSettings") }}</RouterLink>
        </Button>
      </div>
      <div
        v-else-if="!data.supported"
        class="rounded-xl border border-warning/20 bg-warning/5 px-4 py-3 text-sm"
      >
        <span class="font-medium">{{ $t("accounts.notSupported") }}</span>
        <span class="text-muted-foreground"> — {{ $t("accounts.notSupportedHint") }}</span>
      </div>
      <div v-else-if="!data.baselineAt" class="rounded-xl border border-dashed px-4 py-3 text-sm">
        <span class="font-medium">{{ $t("accounts.baselinePending") }}</span>
        <span class="text-muted-foreground"> — {{ $t("accounts.baselinePendingHint") }}</span>
      </div>
      <div
        v-if="data.lastError"
        class="rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3"
      >
        <div class="text-sm font-medium">{{ $t("accounts.error") }}</div>
        <p class="mt-1 break-words text-sm text-muted-foreground">{{ data.lastError }}</p>
      </div>

      <!-- Findings awaiting review -->
      <section class="space-y-2">
        <h2 class="text-base font-semibold">
          {{ $t("accounts.openFindings") }}
          <span v-if="data.openFindings.length" class="text-muted-foreground">
            ({{ data.openFindings.length }})
          </span>
        </h2>

        <EmptyState
          v-if="!data.openFindings.length"
          :icon="IconShieldCheck"
          :title="$t('accounts.noFindings')"
          size="sm"
        />

        <div
          v-for="finding in data.openFindings"
          v-else
          :key="finding.id"
          :class="[
            'rounded-xl border px-4 py-3',
            finding.level === 'red'
              ? 'border-destructive/20 bg-destructive/5'
              : 'border-warning/20 bg-warning/5',
          ]"
        >
          <div class="flex items-start gap-3">
            <StatusIcon :status="finding.level" class="mt-0.5 shrink-0" />
            <div class="min-w-0 flex-1">
              <div class="text-sm font-medium">{{ translateAccountChange(finding) }}</div>
              <AccountFacts :entry="finding" class="mt-2" />
            </div>
            <div class="flex shrink-0 items-center gap-1">
              <Button
                v-if="finding.adminUrl"
                as-child
                variant="ghost"
                size="icon"
                class="size-7"
                :title="$t('accounts.openInShopware')"
              >
                <a :href="finding.adminUrl" target="_blank" rel="noopener noreferrer">
                  <icon-fa6-solid:arrow-up-right-from-square class="size-3" />
                </a>
              </Button>
              <Button variant="outline" size="sm" @click="openAcknowledge(finding)">
                {{ $t("accounts.acknowledge") }}
              </Button>
            </div>
          </div>
        </div>
      </section>

      <!-- Hints: observations about existing accounts, never an alert -->
      <section v-if="data.postureHints.length" class="space-y-2">
        <h2 class="text-base font-semibold">{{ $t("accounts.postureTitle") }}</h2>
        <div class="space-y-2">
          <div
            v-for="(hint, index) in data.postureHints"
            :key="`${hint.hint}-${hint.label}-${index}`"
            class="flex items-start gap-3 rounded-xl border px-4 py-3"
          >
            <icon-fa6-solid:lightbulb class="mt-1 size-3.5 shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <div class="text-sm">{{ translateAccountHint(hint) }}</div>
              <dl class="mt-1.5 flex flex-wrap gap-x-5 gap-y-1 text-xs text-muted-foreground">
                <Fact
                  :label="$t('accounts.facts.kind')"
                  :value="$t('accounts.kind.' + hint.kind)"
                />
                <Fact :label="$t('accounts.facts.account')" :value="hint.label" mono />
                <Fact
                  v-if="hint.lastUsageAt"
                  :label="$t('accounts.facts.lastUsed')"
                  :value="formatDate(hint.lastUsageAt)"
                />
                <Fact
                  v-else-if="hint.hint === 'unusedAdminIntegration'"
                  :label="$t('accounts.facts.lastUsed')"
                  :value="$t('accounts.facts.never')"
                />
                <Fact
                  v-if="hint.accessKeyCount !== undefined"
                  :label="$t('accounts.facts.apiKeys')"
                  :value="String(hint.accessKeyCount)"
                />
                <Fact
                  v-if="hint.createdAt"
                  :label="$t('accounts.facts.created')"
                  :value="formatDate(hint.createdAt)"
                />
              </dl>
            </div>
            <Button
              v-if="hint.adminUrl"
              as-child
              variant="ghost"
              size="icon"
              class="size-7 shrink-0"
              :title="$t('accounts.openInShopware')"
            >
              <a :href="hint.adminUrl" target="_blank" rel="noopener noreferrer">
                <icon-fa6-solid:arrow-up-right-from-square class="size-3" />
              </a>
            </Button>
          </div>
        </div>
      </section>

      <!-- History -->
      <section class="space-y-2">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <h2 class="text-base font-semibold">{{ $t("accounts.history") }}</h2>
          <div class="flex items-center gap-3">
            <span v-if="data.lastSyncedAt" class="text-xs text-muted-foreground">
              {{ $t("accounts.lastSynced", { time: formatDateTime(data.lastSyncedAt) }) }}
            </span>
            <div class="flex flex-wrap gap-1 rounded-lg border bg-muted/50 p-1">
              <button
                v-for="f in filters"
                :key="f.value"
                :class="[
                  'rounded-md px-3 py-1 text-sm font-medium transition-colors',
                  historyFilter === f.value
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground',
                ]"
                @click="historyFilter = f.value"
              >
                {{ f.label }}
              </button>
            </div>
          </div>
        </div>

        <EmptyState
          v-if="!filteredEvents.length"
          :icon="IconClock"
          :title="$t('accounts.noHistory')"
          size="sm"
        />

        <div v-else class="space-y-2">
          <div
            v-for="event in filteredEvents"
            :key="event.id"
            :class="[
              'flex items-start gap-3 rounded-xl border px-4 py-3',
              event.acknowledgedAt ? 'opacity-60' : '',
            ]"
          >
            <StatusIcon
              v-if="event.level !== 'info'"
              :status="event.level"
              class="mt-0.5 shrink-0"
            />
            <icon-fa6-solid:circle-info
              v-else
              class="mt-1 size-3.5 shrink-0 text-muted-foreground"
            />

            <div class="min-w-0 flex-1">
              <div class="text-sm">{{ translateAccountChange(event) }}</div>
              <AccountFacts :entry="event" :detected="event.detectedAt" class="mt-1.5" />
              <div
                v-if="event.acknowledgedAt"
                class="mt-1 flex flex-wrap items-center gap-x-2 text-xs text-muted-foreground"
              >
                <icon-fa6-solid:check class="size-2.5 text-success" />
                <span v-if="event.acknowledgedBy">
                  {{ $t("accounts.acknowledgedBy", { name: event.acknowledgedBy }) }}
                </span>
                <span v-else>
                  {{
                    $t("accounts.acknowledgedAt", { time: formatDateTime(event.acknowledgedAt) })
                  }}
                </span>
                <span v-if="event.acknowledgeNote" class="italic">
                  &ldquo;{{ event.acknowledgeNote }}&rdquo;
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>
    </template>

    <Dialog v-model:open="acknowledgeOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("accounts.acknowledgeTitle") }}</DialogTitle>
          <DialogDescription>{{ $t("accounts.acknowledgeBody") }}</DialogDescription>
        </DialogHeader>

        <div v-if="pending" class="rounded-lg border bg-muted/30 px-3 py-2">
          <div class="text-sm">{{ translateAccountChange(pending) }}</div>
          <AccountFacts :entry="pending" class="mt-2" />
        </div>

        <div class="space-y-2">
          <Label for="acknowledge-note">{{ $t("accounts.acknowledgeNote") }}</Label>
          <Textarea
            id="acknowledge-note"
            v-model="note"
            :maxlength="500"
            :placeholder="$t('accounts.acknowledgeNotePlaceholder')"
          />
        </div>

        <DialogFooter>
          <Button variant="outline" :disabled="submitting" @click="acknowledgeOpen = false">
            {{ $t("common.cancel") }}
          </Button>
          <Button :disabled="submitting" @click="submitAcknowledge">
            {{ $t("accounts.acknowledgeSubmit") }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";

import {
  acknowledgeEnvironmentAccountEvent,
  getEnvironmentAccounts,
  type AccountFinding,
  type EnvironmentAccounts,
} from "@/api/generated";
import { apiErrorMessage, requireMutation } from "@/helpers/apiError";
import { formatDate, formatDateTime } from "@/helpers/formatter";
import { translateAccountChange, translateAccountHint } from "@/helpers/i18n";
import { useAlert } from "@/composables/useAlert";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import EmptyState from "@/components/EmptyState.vue";
import StatCard from "@/components/StatCard.vue";
import StatusIcon from "@/components/StatusIcon.vue";
import AccountFacts from "@/components/accounts/AccountFacts.vue";
import Fact from "@/components/accounts/Fact.vue";
import IconClock from "~icons/fa6-solid/clock-rotate-left";
import IconCircleXmark from "~icons/fa6-solid/circle-xmark";
import IconCircleInfo from "~icons/fa6-solid/circle-info";
import IconLightbulb from "~icons/fa6-solid/lightbulb";
import IconUserShield from "~icons/fa6-solid/user-shield";
import IconShieldCheck from "~icons/fa6-solid/shield-halved";

const route = useRoute();
const { t } = useI18n();
const { info, error } = useAlert();
const { environment, loadEnvironment } = useEnvironmentDetail();

const data = ref<EnvironmentAccounts | null>(null);
const loading = ref(true);
const submitting = ref(false);
const acknowledgeOpen = ref(false);
const pending = ref<AccountFinding | null>(null);
const note = ref("");
const historyFilter = ref<"all" | "findings" | "info">("all");
const loadFailed = ref(false);

const filters = [
  { label: t("common.all"), value: "all" as const },
  { label: t("accounts.filterFindings"), value: "findings" as const },
  { label: t("accounts.filterInfo"), value: "info" as const },
];

function environmentId(): number {
  return Number(route.params.environmentId);
}

const settingsRoute = computed(() => ({
  name: "account.environments.edit",
  params: { environmentId: environmentId() },
}));

const counts = computed(() => {
  const findings = data.value?.openFindings ?? [];
  return {
    red: findings.filter((f) => f.level === "red").length,
    yellow: findings.filter((f) => f.level === "yellow").length,
  };
});

const filteredEvents = computed(() => {
  const events = data.value?.events ?? [];
  if (historyFilter.value === "findings") return events.filter((e) => e.level !== "info");
  if (historyFilter.value === "info") return events.filter((e) => e.level === "info");
  return events;
});

async function load() {
  const requestedId = environmentId();
  loading.value = true;
  loadFailed.value = false;
  try {
    // The generated client reports failures through `error` and does not
    // throw, so ignoring that field would leave a blank tab on a server or
    // network error.
    const { data: result, error: apiError } = await getEnvironmentAccounts({
      path: { environmentId: requestedId },
    });

    // The layout reuses this page across environments, so a slower request for
    // the previous one can land last. Dropping late responses keeps it from
    // replacing the current findings -- and from copying a stale count into
    // the badge.
    if (requestedId !== environmentId()) return;

    if (apiError) {
      loadFailed.value = true;
      error(apiErrorMessage(apiError, t("accounts.loadFailed")));
      return;
    }

    data.value = result ?? null;
    syncBadge();
  } catch (err) {
    loadFailed.value = true;
    error(err instanceof Error ? err.message : String(err));
  } finally {
    if (requestedId === environmentId()) loading.value = false;
  }
}

function openAcknowledge(finding: AccountFinding) {
  pending.value = finding;
  note.value = "";
  acknowledgeOpen.value = true;
}

async function submitAcknowledge() {
  const finding = pending.value;
  if (!finding) return;

  submitting.value = true;
  try {
    requireMutation(
      await acknowledgeEnvironmentAccountEvent({
        path: { environmentId: environmentId(), eventId: finding.id },
        body: { note: note.value.trim() || undefined },
      }),
      t("accounts.acknowledgeFailed"),
    );
    acknowledgeOpen.value = false;
    pending.value = null;
    await load();
    // The backend queues a scrape to recompute the status; reload the shared
    // environment so the header and the tab badge follow it instead of keeping
    // the stale colour for the rest of the visit.
    await loadEnvironment(true);
    info(t("accounts.acknowledged"));
  } catch (err) {
    error(err instanceof Error ? err.message : t("accounts.acknowledgeFailed"));
  } finally {
    submitting.value = false;
  }
}

// Keep the tab badge in step with what this view just changed, without
// re-fetching the whole environment.
function syncBadge() {
  if (!environment.value || !data.value) return;
  environment.value = {
    ...environment.value,
    openAccountFindings: data.value.openFindings.length,
    accountMonitoringEnabled: data.value.enabled,
  };
}

watch(() => route.params.environmentId, load, { immediate: true });
</script>
