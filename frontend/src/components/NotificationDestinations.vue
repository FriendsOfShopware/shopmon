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
      <p v-if="!destinations.length" class="text-sm text-muted-foreground">
        {{ $t("destinations.empty") }}
      </p>
      <ul v-else class="space-y-3">
        <li
          v-for="destination in destinations"
          :key="destination.id"
          class="flex flex-wrap items-center gap-3 rounded-lg border p-3"
        >
          <component
            :is="providers[destination.provider]?.icon ?? IconBell"
            class="size-5 shrink-0"
          />
          <div class="min-w-0 flex-1">
            <p class="break-words text-sm font-medium">{{ destination.name }}</p>
            <p class="break-all text-xs text-muted-foreground">
              {{ providers[destination.provider]?.label ?? destination.provider }} ·
              {{
                destination.summary === "Incoming webhook"
                  ? $t("destinations.webhook")
                  : destination.summary
              }}
            </p>
          </div>
          <Badge class="text-xs">{{ $t("destinations.configured") }}</Badge>
          <div v-if="canManage" class="flex flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="isBusy"
              @click="sendTest(destination.id)"
              >{{ $t("destinations.test") }}</Button
            >
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="isBusy"
              @click="edit(destination)"
              >{{ $t("destinations.edit") }}</Button
            >
            <Button
              type="button"
              variant="ghost"
              size="sm"
              :disabled="isBusy"
              @click="remove(destination.id)"
              >{{ $t("destinations.remove") }}</Button
            >
          </div>
        </li>
      </ul>
      <template v-if="canManage">
        <Button v-if="!formOpen" type="button" variant="outline" size="sm" @click="add">{{
          $t("destinations.add")
        }}</Button>
        <form v-else class="space-y-3 rounded-lg border p-4" @submit.prevent="save">
          <p class="text-sm font-medium">
            {{ editingId ? $t("destinations.edit") : $t("destinations.add") }}
          </p>
          <label class="block space-y-1 text-sm">
            <span>{{ $t("destinations.name") }}</span>
            <Input
              v-model="name"
              :disabled="isBusy"
              required
              maxlength="100"
              :placeholder="$t('destinations.namePlaceholder')"
            />
          </label>
          <label class="block space-y-1 text-sm">
            <span>{{ $t("destinations.dsn") }}</span>
            <Input
              v-model="dsn"
              :disabled="isBusy"
              type="password"
              autocomplete="new-password"
              spellcheck="false"
              maxlength="8192"
              :required="!editingId"
              :placeholder="editingId ? $t('destinations.keep') : 'slack://default?webhook_url=…'"
            />
          </label>
          <p v-if="detectedProvider" class="flex items-center gap-2 text-xs text-muted-foreground">
            <component :is="detectedProvider.icon" class="size-4" />{{ detectedProvider.label }}
          </p>
          <p class="text-xs text-muted-foreground">{{ $t("destinations.secretHelp") }}</p>
          <details class="text-xs text-muted-foreground">
            <summary class="cursor-pointer">{{ $t("destinations.examples") }}</summary>
            <ul class="mt-2 space-y-2">
              <li v-for="example in examples" :key="example.label">
                <span class="font-medium">{{ example.label }}</span
                ><code class="mt-1 block break-all">{{ example.dsn }}</code>
              </li>
            </ul>
            <p class="mt-2">{{ $t("destinations.encodingHelp") }}</p>
          </details>
          <div class="flex gap-2">
            <Button
              type="submit"
              size="sm"
              :disabled="isBusy || !name.trim() || (!editingId && !dsn.trim())"
              >{{ $t("common.save") }}</Button
            >
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
import { computed, onMounted, ref } from "vue";
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
const providers: Record<string, { label: string; icon: typeof IconBell }> = {
  slack: { label: "Slack", icon: IconSlack },
  telegram: { label: "Telegram", icon: IconTelegram },
  discord: { label: "Discord", icon: IconDiscord },
  gotify: { label: "Gotify", icon: IconBell },
  microsoftteams: { label: "Microsoft Teams", icon: IconMicrosoft },
};
const examples = [
  {
    label: "Slack · Webhook",
    dsn: "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2FT000%2FB000%2FTOKEN",
  },
  { label: "Slack · Bot", dsn: "slack://xoxb-TOKEN@default?channel=C123" },
  { label: "Telegram", dsn: "telegram://BOT_TOKEN@default?channel=CHAT_ID" },
  { label: "Discord", dsn: "discord://WEBHOOK_TOKEN@default?webhook_id=WEBHOOK_ID" },
  { label: "Gotify", dsn: "gotify://APP_TOKEN@notify.example.com" },
  { label: "Microsoft Teams", dsn: "microsoftteams://WEBHOOK_ID@default?token=TOKEN" },
];
const destinations = ref<NotificationDestination[]>([]);
const canManage = ref(false);
const isLoading = ref(true);
const loadFailed = ref(false);
const isBusy = ref(false);
const formOpen = ref(false);
const editingId = ref("");
const name = ref("");
const dsn = ref("");
const detectedProvider = computed(() => providers[dsn.value.trim().split(":", 1)[0] ?? ""]);
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
  dsn.value = "";
}
function add() {
  cancel();
  formOpen.value = true;
}
function edit(destination: NotificationDestination) {
  cancel();
  editingId.value = destination.id;
  name.value = destination.name;
  formOpen.value = true;
}
async function save() {
  if (isBusy.value) return;
  isBusy.value = true;
  try {
    const body = { name: name.value.trim(), dsn: dsn.value.trim() };
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
