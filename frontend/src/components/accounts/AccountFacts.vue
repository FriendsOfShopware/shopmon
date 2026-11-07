<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { AccountEvent, AccountFinding } from "@/api/generated";
import { formatDate, formatDateTime } from "@/helpers/formatter";
import Fact from "./Fact.vue";

const props = defineProps<{
  entry: AccountFinding | AccountEvent;
  /** When set, the detection time is shown as part of the facts. */
  detected?: string;
}>();

const { t } = useI18n();

type Row = { label: string; value: string; mono?: boolean };

/**
 * The facts a reviewer needs in order to judge a change without following the
 * deep link into the shop. `account` is the account's current state and is
 * absent once the account is gone — the label and the change params are then
 * all that remains, which is why they are listed first.
 */
const rows = computed<Row[]>(() => {
  const e = props.entry;
  const a = e.account;
  const out: Row[] = [
    { label: t("accounts.facts.kind"), value: t(`accounts.kind.${e.kind}`) },
    { label: t("accounts.facts.account"), value: e.label, mono: true },
  ];

  if (props.detected) {
    out.push({ label: t("accounts.facts.detected"), value: formatDateTime(props.detected) });
  }

  // Counts carried by the change itself (role grew from x to y, keys x to y).
  const params = (e.params ?? {}) as Record<string, unknown>;
  if (params.from !== undefined && params.to !== undefined) {
    out.push({
      label: t("accounts.facts.change"),
      value: `${String(params.from)} → ${String(params.to)}`,
    });
  }
  if (typeof params.count === "number") {
    out.push({ label: t("accounts.facts.count"), value: String(params.count) });
  }
  if (typeof params.holders === "number") {
    out.push({ label: t("accounts.facts.holders"), value: String(params.holders) });
  }
  if (typeof params.appName === "string" && params.appName) {
    out.push({ label: t("accounts.facts.app"), value: params.appName });
  }

  if (!a) {
    out.push({ label: t("accounts.facts.state"), value: t("accounts.facts.gone") });
    return out;
  }

  if (e.kind === "acl_role") {
    out.push({ label: t("accounts.facts.privileges"), value: String(a.privilegeCount) });
  } else {
    out.push({
      label: t("accounts.facts.admin"),
      value: a.admin ? t("common.yes") : t("common.no"),
    });
    out.push({
      label: t("accounts.facts.state"),
      value: a.active ? t("common.active") : t("common.inactive"),
    });
    if (a.roleCount > 0) {
      out.push({ label: t("accounts.facts.roles"), value: String(a.roleCount) });
    }
    if (a.accessKeyCount > 0) {
      out.push({ label: t("accounts.facts.apiKeys"), value: String(a.accessKeyCount) });
    }
    if (a.managedByApp) {
      out.push({
        label: t("accounts.facts.app"),
        value: a.appName ?? t("common.yes"),
      });
    }
    if (a.passwordRecoveryPending) {
      out.push({ label: t("accounts.facts.passwordReset"), value: t("accounts.facts.pending") });
    }
    if (a.lastUsageAt) {
      out.push({ label: t("accounts.facts.lastUsed"), value: formatDate(a.lastUsageAt) });
    }
  }

  if (a.createdAt) {
    out.push({ label: t("accounts.facts.created"), value: formatDate(a.createdAt) });
  }
  return out;
});
</script>

<template>
  <dl class="flex flex-wrap gap-x-5 gap-y-1 text-xs">
    <Fact
      v-for="(row, i) in rows"
      :key="i"
      :label="row.label"
      :value="row.value"
      :mono="row.mono"
    />
  </dl>
</template>
