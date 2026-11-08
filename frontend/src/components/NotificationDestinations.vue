<template>
  <CardSection
    :icon="IconBell"
    :title="$t('destinations.title')"
    :description="shopId ? $t('destinations.shopDescription') : $t('destinations.description')"
  >
    <p v-if="isLoading" class="text-sm text-muted-foreground" role="status">
      {{ $t("destinations.loading") }}
    </p>
    <div
      v-else-if="loadFailed"
      class="flex items-center gap-3 text-sm text-destructive"
      role="alert"
    >
      {{ $t("destinations.failedLoad") }}
      <Button variant="outline" size="sm" @click="load">{{ $t("destinations.retry") }}</Button>
    </div>
    <div v-else class="space-y-4">
      <p v-if="!destinations.length && !formOpen" class="text-sm text-muted-foreground">
        {{ $t("destinations.empty") }}
      </p>
      <ul v-else-if="destinations.length" class="space-y-3">
        <li
          v-for="destination in destinations"
          :key="destination.id"
          class="flex flex-wrap items-center gap-3 rounded-lg border p-3"
        >
          <component
            :is="providerDetails(destination.provider)?.icon ?? IconBell"
            class="size-5 shrink-0"
          />
          <div class="min-w-0 flex-1">
            <p class="break-words text-sm font-medium">{{ destination.name }}</p>
            <p v-if="!destination.connectionError" class="break-all text-xs text-muted-foreground">
              {{ providerDetails(destination.provider)?.label ?? destination.provider }} ·
              {{
                destination.summary === "Incoming webhook"
                  ? $t("destinations.webhook")
                  : destination.summary
              }}
            </p>
            <p v-if="destination.connectionError" class="text-xs text-muted-foreground">
              {{ $t("destinations.connectionError") }}
            </p>
          </div>
          <Badge class="text-xs">{{
            $t(
              destination.connectionError
                ? "destinations.connectionErrorLabel"
                : "destinations.configured",
            )
          }}</Badge>
          <div class="flex flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="isBusy"
              :aria-expanded="eventSettingsId === destination.id"
              @click="eventSettingsId = eventSettingsId === destination.id ? '' : destination.id"
              >{{ $t("destinations.eventSettings") }}</Button
            >
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="isBusy"
              v-if="canManage"
              @click="sendTest(destination.id)"
              >{{ $t("destinations.test") }}</Button
            >
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="isBusy"
              v-if="canManage"
              @click="edit(destination)"
              >{{ $t("destinations.edit") }}</Button
            >
            <Button
              type="button"
              variant="ghost"
              size="sm"
              :disabled="isBusy"
              v-if="canManage"
              @click="remove(destination.id)"
              >{{ $t("destinations.remove") }}</Button
            >
          </div>
          <div v-if="eventSettingsId === destination.id" class="w-full space-y-2 border-t pt-3">
            <p class="text-xs text-muted-foreground">
              {{ $t("destinations.defaultPreferenceHint") }}
            </p>
            <NotificationMatrix
              :destination="{ id: destination.id, organizationId }"
              :disabled="!canManage || isBusy"
            />
          </div>
        </li>
      </ul>
      <template v-if="canManage">
        <Button v-if="!formOpen" type="button" variant="outline" size="sm" @click="add">{{
          $t("destinations.add")
        }}</Button>
        <form
          v-else
          class="space-y-5 rounded-lg border bg-muted/20 p-4 sm:p-5"
          @submit.prevent="save"
        >
          <div>
            <p class="text-sm font-semibold">
              {{ editingId ? $t("destinations.edit") : $t("destinations.add") }}
            </p>
            <p class="mt-1 text-sm text-muted-foreground">{{ $t("destinations.setupHint") }}</p>
          </div>
          <label class="block max-w-lg space-y-1.5 text-sm">
            <span class="font-medium">{{ $t("destinations.name") }}</span>
            <Input
              v-model="name"
              :disabled="isBusy"
              required
              maxlength="100"
              :placeholder="$t('destinations.namePlaceholder')"
            />
          </label>
          <div
            v-if="editingId && !replaceConnection"
            class="flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-background p-3"
          >
            <div class="flex items-center gap-3">
              <component :is="providers[connection.provider].icon" class="size-5" />
              <div>
                <p class="text-sm font-medium">{{ providers[connection.provider].label }}</p>
                <p class="text-xs text-muted-foreground">{{ $t("destinations.keepConnection") }}</p>
              </div>
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="isBusy"
              @click="replaceConnection = true"
              >{{ $t("destinations.changeConnection") }}</Button
            >
          </div>
          <template v-if="!editingId || replaceConnection">
            <fieldset :disabled="isBusy" class="space-y-2">
              <legend class="mb-2 text-sm font-medium">{{ $t("destinations.provider") }}</legend>
              <div class="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-5">
                <label
                  v-for="(provider, key) in providers"
                  :key="key"
                  class="flex cursor-pointer items-center gap-2 rounded-lg border bg-background p-3 text-sm transition-colors focus-within:ring-2 focus-within:ring-ring"
                  :class="
                    connection.provider === key
                      ? 'border-primary bg-primary/5 text-primary'
                      : 'hover:bg-muted'
                  "
                >
                  <input
                    type="radio"
                    class="sr-only"
                    :name="providerGroup"
                    :value="key"
                    :checked="connection.provider === key"
                    @change="chooseProvider(key)"
                  />
                  <component :is="provider.icon" class="size-5 shrink-0" /><span>{{
                    provider.label
                  }}</span>
                </label>
              </div>
            </fieldset>
            <div class="max-w-lg space-y-4">
              <label v-if="connection.provider === 'slack'" class="block space-y-1.5 text-sm">
                <span class="font-medium">{{ $t("destinations.connectionMethod") }}</span>
                <select
                  v-model="connection.slackMode"
                  :disabled="isBusy"
                  class="flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  @change="clearCredentials"
                >
                  <option value="webhook">{{ $t("destinations.slackWebhookMethod") }}</option>
                  <option value="bot">{{ $t("destinations.slackBotMethod") }}</option>
                </select>
              </label>
              <label v-if="usesWebhook" class="block space-y-1.5 text-sm">
                <span class="font-medium">{{ $t("destinations.webhookUrl") }}</span>
                <Input
                  v-model="connection.webhookUrl"
                  :disabled="isBusy"
                  type="password"
                  autocomplete="new-password"
                  spellcheck="false"
                  required
                  maxlength="8192"
                  :placeholder="webhookPlaceholder"
                />
                <span class="block text-xs text-muted-foreground">{{
                  $t(`destinations.help.${connection.provider}`)
                }}</span>
              </label>
              <label v-if="connection.provider === 'gotify'" class="block space-y-1.5 text-sm">
                <span class="font-medium">{{ $t("destinations.serverUrl") }}</span>
                <Input
                  v-model="connection.serverUrl"
                  :disabled="isBusy"
                  type="url"
                  required
                  placeholder="https://notify.example.com"
                />
                <span class="block text-xs text-muted-foreground">{{
                  $t("destinations.help.gotify")
                }}</span>
              </label>
              <label v-if="!usesWebhook" class="block space-y-1.5 text-sm">
                <span class="font-medium">{{
                  connection.provider === "gotify"
                    ? $t("destinations.appToken")
                    : $t("destinations.botToken")
                }}</span>
                <Input
                  v-model="connection.token"
                  :disabled="isBusy"
                  type="password"
                  autocomplete="new-password"
                  spellcheck="false"
                  required
                  maxlength="4096"
                  :placeholder="
                    connection.provider === 'slack' ? 'xoxb-…' : $t('destinations.tokenPlaceholder')
                  "
                />
                <span
                  v-if="connection.provider === 'telegram'"
                  class="block text-xs text-muted-foreground"
                  >{{ $t("destinations.help.telegram") }}</span
                >
              </label>
              <label
                v-if="
                  connection.provider === 'telegram' ||
                  (connection.provider === 'slack' && connection.slackMode === 'bot')
                "
                class="block space-y-1.5 text-sm"
              >
                <span class="font-medium">{{
                  connection.provider === "telegram"
                    ? $t("destinations.chatId")
                    : $t("destinations.channelId")
                }}</span>
                <Input
                  v-model="connection.channel"
                  :disabled="isBusy"
                  required
                  :placeholder="
                    connection.provider === 'telegram' ? '-1001234567890' : 'C0123456789'
                  "
                />
                <span class="block text-xs text-muted-foreground">{{
                  connection.provider === "telegram"
                    ? $t("destinations.chatHelp")
                    : $t("destinations.channelHelp")
                }}</span>
              </label>
            </div>
            <p class="text-xs text-muted-foreground">{{ $t("destinations.credentialsHelp") }}</p>
          </template>
          <p v-if="!editingId" class="text-xs text-muted-foreground">
            {{ $t("destinations.newPreferenceHint") }}
          </p>
          <p v-if="formError" role="alert" class="text-sm text-destructive">{{ formError }}</p>
          <div class="flex items-center gap-2 border-t pt-4">
            <Button type="submit" size="sm" :disabled="isBusy || !name.trim()">{{
              isBusy ? $t("destinations.saving") : $t("common.save")
            }}</Button>
            <Button type="button" variant="ghost" size="sm" :disabled="isBusy" @click="cancel">{{
              $t("destinations.cancel")
            }}</Button>
          </div>
        </form>
      </template>
      <p v-else class="text-xs text-muted-foreground">{{ $t("destinations.readOnly") }}</p>
    </div>
  </CardSection>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, useId, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  listNotificationDestinations,
  listShopNotificationDestinations,
  createNotificationDestination,
  createShopNotificationDestination,
  updateNotificationDestination,
  updateShopNotificationDestination,
  deleteNotificationDestination,
  deleteShopNotificationDestination,
  testNotificationDestination,
  testShopNotificationDestination,
  type NotificationDestination,
} from "@/api/generated";
import { useAlert } from "@/composables/useAlert";
import {
  buildDestinationDSN,
  emptyConnection,
  type DestinationProvider,
} from "@/helpers/notificationDestination";
import NotificationMatrix from "@/components/NotificationMatrix.vue";
import CardSection from "@/components/CardSection.vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import IconBell from "~icons/fa6-solid/bell";
import IconSlack from "~icons/fa6-brands/slack";
import IconTelegram from "~icons/fa6-brands/telegram";
import IconDiscord from "~icons/fa6-brands/discord";
import IconMicrosoft from "~icons/fa6-brands/microsoft";

const props = defineProps<{ organizationId: string; shopId?: number }>();
const { t } = useI18n();
const alert = useAlert();
const providers: Record<DestinationProvider, { label: string; icon: typeof IconBell }> = {
  slack: { label: "Slack", icon: IconSlack },
  telegram: { label: "Telegram", icon: IconTelegram },
  discord: { label: "Discord", icon: IconDiscord },
  gotify: { label: "Gotify", icon: IconBell },
  microsoftteams: { label: "Microsoft Teams", icon: IconMicrosoft },
};
function providerDetails(provider: string) {
  return providers[provider as DestinationProvider];
}
const providerGroup = useId();
const destinations = ref<NotificationDestination[]>([]);
const canManage = ref(false);
const isLoading = ref(true);
const loadFailed = ref(false);
const isBusy = ref(false);
const formOpen = ref(false);
const editingId = ref("");
const eventSettingsId = ref("");
const name = ref("");
const connection = ref(emptyConnection());
const replaceConnection = ref(false);
const formError = ref("");
const usesWebhook = computed(
  () =>
    connection.value.provider === "discord" ||
    connection.value.provider === "microsoftteams" ||
    (connection.value.provider === "slack" && connection.value.slackMode === "webhook"),
);
const webhookPlaceholder = computed(
  () =>
    ({
      slack: "https://hooks.slack.com/services/…",
      discord: "https://discord.com/api/webhooks/…",
      microsoftteams: "https://outlook.office.com/webhook/…",
    })[connection.value.provider as "slack" | "discord" | "microsoftteams"],
);
watch(
  connection,
  () => {
    formError.value = "";
  },
  { deep: true, flush: "sync" },
);
function clearCredentials() {
  connection.value = {
    ...emptyConnection(connection.value.provider),
    slackMode: connection.value.slackMode,
  };
}
function chooseProvider(provider: DestinationProvider) {
  if (provider !== connection.value.provider) connection.value = emptyConnection(provider);
}
const orgPath = () => ({ orgId: props.organizationId });
const shopPath = () => ({ ...orgPath(), shopId: props.shopId! });

async function load() {
  isLoading.value = true;
  loadFailed.value = false;
  try {
    const { data, error } = props.shopId
      ? await listShopNotificationDestinations({ path: shopPath() })
      : await listNotificationDestinations({ path: orgPath() });
    if (error || !data) {
      loadFailed.value = true;
      return;
    }
    destinations.value = data.destinations;
    canManage.value = data.canManage;
  } catch {
    loadFailed.value = true;
  } finally {
    isLoading.value = false;
  }
}
function cancel() {
  formOpen.value = false;
  editingId.value = "";
  name.value = "";
  connection.value = emptyConnection();
  replaceConnection.value = false;
  formError.value = "";
}
function add() {
  cancel();
  formOpen.value = true;
}
function edit(destination: NotificationDestination) {
  cancel();
  editingId.value = destination.id;
  name.value = destination.name;
  replaceConnection.value = destination.connectionError;
  connection.value = emptyConnection(
    providerDetails(destination.provider) ? (destination.provider as DestinationProvider) : "slack",
  );
  if (destination.provider === "slack" && destination.summary !== "Incoming webhook")
    connection.value.slackMode = "bot";
  if (
    destination.provider === "telegram" ||
    (destination.provider === "slack" && connection.value.slackMode === "bot")
  )
    connection.value.channel = destination.summary;
  formOpen.value = true;
}
async function save() {
  if (isBusy.value) return;
  let dsn = "";
  if (!editingId.value || replaceConnection.value) {
    try {
      dsn = buildDestinationDSN(connection.value);
    } catch (error) {
      const key = error instanceof Error ? error.message : "invalidConnection";
      formError.value = t(
        `destinations.validation.${key === "URI malformed" ? "invalidConnection" : key}`,
      );
      return;
    }
  }
  isBusy.value = true;
  try {
    const body = { name: name.value.trim(), dsn };
    const result = editingId.value
      ? props.shopId
        ? await updateShopNotificationDestination({
            path: { ...shopPath(), destinationId: editingId.value },
            body,
          })
        : await updateNotificationDestination({
            path: { ...orgPath(), destinationId: editingId.value },
            body,
          })
      : props.shopId
        ? await createShopNotificationDestination({ path: shopPath(), body })
        : await createNotificationDestination({ path: orgPath(), body });
    if (result.error) {
      alert.error(t("destinations.failedSave"));
      return;
    }
    cancel();
    alert.success(t("destinations.saved"));
    await load();
  } catch {
    alert.error(t("destinations.failedSave"));
  } finally {
    isBusy.value = false;
  }
}
async function sendTest(destinationId: string) {
  if (isBusy.value) return;
  isBusy.value = true;
  try {
    const { error } = props.shopId
      ? await testShopNotificationDestination({ path: { ...shopPath(), destinationId } })
      : await testNotificationDestination({ path: { ...orgPath(), destinationId } });
    if (error) alert.error(t("destinations.failedTest"));
    else alert.success(t("destinations.testSent"));
  } catch {
    alert.error(t("destinations.failedTest"));
  } finally {
    isBusy.value = false;
  }
}
async function remove(destinationId: string) {
  if (isBusy.value) return;
  isBusy.value = true;
  try {
    const { error } = props.shopId
      ? await deleteShopNotificationDestination({ path: { ...shopPath(), destinationId } })
      : await deleteNotificationDestination({ path: { ...orgPath(), destinationId } });
    if (error) {
      alert.error(t("destinations.failedRemove"));
      return;
    }
    if (editingId.value === destinationId) cancel();
    alert.success(t("destinations.removed"));
    await load();
  } catch {
    alert.error(t("destinations.failedRemove"));
  } finally {
    isBusy.value = false;
  }
}
onMounted(load);
</script>
