<template>
  <CardSection
    :icon="IconSlack"
    :title="$t('slack.title')"
    :description="shopId ? $t('slack.shopDescription') : $t('slack.description')"
  >
    <template #action>
      <Badge v-if="configured" class="bg-success/10 text-success border-success/30 text-xs">
        {{ $t("slack.connected") }}
      </Badge>
    </template>

    <div v-if="isLoading" class="flex items-center gap-2 text-sm text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-4" />
    </div>

    <div v-else-if="loadFailed" class="flex items-center gap-3 text-sm text-destructive">
      {{ $t("slack.failedLoad") }}
      <Button type="button" variant="outline" size="sm" @click="load">
        {{ $t("slack.retry") }}
      </Button>
    </div>

    <p v-else-if="!canManage" class="text-sm text-muted-foreground">
      {{ configured ? $t("slack.connectedHint") : $t("slack.notConnectedMember") }}
    </p>

    <form v-else class="space-y-3" @submit.prevent="save">
      <div class="flex flex-col gap-2 sm:flex-row">
        <Input
          v-model="webhookUrl"
          type="password"
          autocomplete="off"
          class="flex-1"
          :placeholder="configured ? $t('slack.webhookKeep') : 'https://hooks.slack.com/services/…'"
        />
        <Button type="submit" size="sm" :disabled="!webhookUrl || isBusy">
          <icon-line-md:loading-twotone-loop v-if="isSaving" class="mr-1.5 size-3.5" />
          {{ $t("common.save") }}
        </Button>
      </div>
      <p class="text-xs text-muted-foreground">{{ $t("slack.webhookHelp") }}</p>

      <div v-if="configured" class="flex gap-2">
        <Button type="button" variant="outline" size="sm" :disabled="isBusy" @click="sendTest">
          <icon-fa6-solid:paper-plane v-if="!isTesting" class="mr-1.5 size-3" />
          <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3" />
          {{ $t("slack.sendTest") }}
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          class="text-muted-foreground hover:text-destructive"
          :disabled="isBusy"
          @click="remove"
        >
          <icon-fa6-solid:trash class="mr-1.5 size-3" />
          {{ $t("slack.disconnect") }}
        </Button>
      </div>
    </form>
  </CardSection>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import {
  deleteShopSlackWebhook,
  deleteSlackWebhook,
  getShopSlackWebhook,
  getSlackWebhook,
  testShopSlackWebhook,
  testSlackWebhook,
  updateShopSlackWebhook,
  updateSlackWebhook,
} from "@/api/generated";
import { useAlert } from "@/composables/useAlert";
import CardSection from "@/components/CardSection.vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import IconSlack from "~icons/fa6-brands/slack";

// Without shopId the card manages the organization webhook, with it the
// shop's. Whether the viewer may manage it comes from the API.
const props = defineProps<{
  organizationId: string;
  shopId?: number;
}>();

type Result = { error?: unknown };
const api = {
  get: () =>
    props.shopId
      ? getShopSlackWebhook({ path: { orgId: props.organizationId, shopId: props.shopId } })
      : getSlackWebhook({ path: { orgId: props.organizationId } }),
  update: (webhookUrl: string): Promise<Result> =>
    props.shopId
      ? updateShopSlackWebhook({
          path: { orgId: props.organizationId, shopId: props.shopId },
          body: { webhookUrl },
        })
      : updateSlackWebhook({ path: { orgId: props.organizationId }, body: { webhookUrl } }),
  test: (): Promise<Result> =>
    props.shopId
      ? testShopSlackWebhook({ path: { orgId: props.organizationId, shopId: props.shopId } })
      : testSlackWebhook({ path: { orgId: props.organizationId } }),
  remove: (): Promise<Result> =>
    props.shopId
      ? deleteShopSlackWebhook({ path: { orgId: props.organizationId, shopId: props.shopId } })
      : deleteSlackWebhook({ path: { orgId: props.organizationId } }),
};

const { t } = useI18n();
const alert = useAlert();

const configured = ref(false);
const canManage = ref(false);
const webhookUrl = ref("");
const isLoading = ref(true);
const isSaving = ref(false);
const isTesting = ref(false);
const isRemoving = ref(false);
const loadFailed = ref(false);
// One request at a time: each handler sets `configured` from its own response,
// so overlapping save and disconnect could leave the card showing the wrong state.
const isBusy = computed(() => isSaving.value || isTesting.value || isRemoving.value);

function errorMessage(error: unknown, fallback: string): string {
  return (error as { message?: string } | undefined)?.message ?? fallback;
}

async function load() {
  isLoading.value = true;
  loadFailed.value = false;
  try {
    const { data, error } = await api.get();
    // A failed load must not read as "not connected", which would also hide
    // the controls from managers.
    if (error || !data) {
      loadFailed.value = true;
      return;
    }
    configured.value = data.configured;
    canManage.value = data.canManage;
  } catch {
    loadFailed.value = true;
  } finally {
    isLoading.value = false;
  }
}

async function save() {
  isSaving.value = true;
  try {
    const { error } = await api.update(webhookUrl.value.trim());
    if (error) {
      alert.error(errorMessage(error, t("slack.failedSave")));
      return;
    }
    webhookUrl.value = "";
    configured.value = true;
    alert.success(t("slack.saved"));
  } finally {
    isSaving.value = false;
  }
}

async function sendTest() {
  isTesting.value = true;
  try {
    const { error } = await api.test();
    if (error) {
      alert.error(errorMessage(error, t("slack.failedTest")));
      return;
    }
    alert.success(t("slack.testSent"));
  } finally {
    isTesting.value = false;
  }
}

async function remove() {
  isRemoving.value = true;
  try {
    const { error } = await api.remove();
    if (error) {
      alert.error(errorMessage(error, t("slack.failedDisconnect")));
      return;
    }
    configured.value = false;
    alert.success(t("slack.disconnected"));
  } finally {
    isRemoving.value = false;
  }
}

onMounted(load);
</script>
