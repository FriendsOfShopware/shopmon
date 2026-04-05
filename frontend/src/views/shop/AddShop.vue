<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("shop.newShop") }}</h1>

    <form v-if="!isLoadingOrgs" @submit="onSubmit" class="space-y-6">
      <!-- Shop details card -->
      <Card>
        <CardHeader class="pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:folder class="size-4 text-muted-foreground" />
            {{ $t("shop.shopInfo") }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="grid gap-4 sm:grid-cols-2">
            <FormField v-slot="{ componentField }" name="name">
              <FormItem>
                <FormLabel>{{ $t("common.name") }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="organizationId">
              <FormItem>
                <FormLabel>{{ $t("settings.organization") }}</FormLabel>
                <Select v-bind="componentField">
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue :placeholder="$t('shop.selectOrganization')" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem v-for="org in organizations" :key="org.id" :value="org.id">
                      {{ org.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="description" class="sm:col-span-2">
              <FormItem>
                <FormLabel>{{ $t("common.description") }}</FormLabel>
                <FormControl>
                  <Textarea v-bind="componentField" :placeholder="$t('shop.optionalDescription')" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="gitUrl" class="sm:col-span-2">
              <FormItem>
                <FormLabel>{{ $t("shop.gitRepoUrl") }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="url"
                    placeholder="https://github.com/org/repo"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>
        </CardContent>
      </Card>

      <!-- First environment card -->
      <Card>
        <CardHeader class="pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:earth-americas class="size-4 text-muted-foreground" />
            {{ $t("shop.firstEnvironment") }}
          </CardTitle>
          <p class="text-sm text-muted-foreground">
            {{ $t("shop.firstEnvironmentDesc") }}
          </p>
        </CardHeader>
        <CardContent class="space-y-6">
          <div class="grid gap-4 sm:grid-cols-2">
            <FormField v-slot="{ componentField }" name="environmentName">
              <FormItem>
                <FormLabel>{{ $t("environment.environmentName") }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" placeholder="Production" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="environmentUrl">
              <FormItem>
                <FormLabel>{{ $t("common.url") }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="url"
                    placeholder="https://my-shop.example.com"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>

          <Separator />

          <!-- Integration -->
          <div>
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-sm font-semibold">{{ $t("environment.integration") }}</h3>
                <p class="mt-0.5 text-xs text-muted-foreground">
                  <i18n-t keypath="environment.integrationDesc" tag="span">
                    <template #pluginLink>
                      <a
                        href="https://github.com/FriendsOfShopware/FroshShopmon"
                        target="_blank"
                        class="text-primary hover:underline"
                        >{{ $t("environment.pluginName") }}</a
                      >
                    </template>
                  </i18n-t>
                </p>
              </div>
              <Button type="button" variant="outline" size="sm" @click="openPluginModal">
                <icon-fa6-solid:plug class="mr-1.5 size-3" />
                {{ $t("environment.connectPlugin") }}
              </Button>
            </div>

            <div class="mt-4 grid gap-4 sm:grid-cols-2">
              <FormField v-slot="{ componentField }" name="clientId">
                <FormItem>
                  <FormLabel>{{ $t("environment.clientId") }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="clientSecret">
                <FormItem>
                  <FormLabel>{{ $t("environment.clientSecret") }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" type="password" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Submit -->
      <div class="flex justify-end">
        <Button :disabled="isSubmitting" type="submit">
          <icon-fa6-solid:floppy-disk
            v-if="!isSubmitting"
            class="mr-1.5 size-3.5"
            aria-hidden="true"
          />
          <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
          {{ $t("shop.createShop") }}
        </Button>
      </div>
    </form>

    <!-- Plugin Connection Modal -->
    <PluginConnectionModal
      v-model:base64="pluginBase64"
      :show="showPluginModal"
      :error="pluginError"
      @close="closePluginModal"
      @import="processPluginData"
    />
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import PluginConnectionModal from "@/components/modal/PluginConnectionModal.vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

const { t } = useI18n();
const { error } = useAlert();
const router = useRouter();
const route = useRoute();

interface Organization {
  id: string;
  name: string;
}

const organizations = ref<Organization[]>([]);
const isLoadingOrgs = ref(true);

const showPluginModal = ref(false);
const pluginBase64 = ref("");
const pluginError = ref("");

async function loadOrganizations() {
  isLoadingOrgs.value = true;
  try {
    const { data } = await api.GET("/auth/list-organizations");
    if (data) {
      organizations.value = data;
    }
  } catch {
    // silently ignore
  } finally {
    isLoadingOrgs.value = false;
  }
}

const isValidUrl = (url: string) => {
  try {
    return !!new URL(url);
  } catch {
    return false;
  }
};

const validationSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.shopNameRequired")),
    description: z.string().optional(),
    gitUrl: z.string().url(t("validation.urlInvalid")).optional().or(z.literal("")),
    organizationId: z.string().min(1, t("validation.orgRequired")),
    environmentName: z.string().min(1, t("validation.environmentNameRequired")),
    environmentUrl: z
      .string()
      .min(1, t("validation.environmentUrlRequired"))
      .refine((val) => isValidUrl(val), t("validation.environmentUrlInvalid")),
    clientId: z.string().min(1, t("validation.clientIdRequired")),
    clientSecret: z.string().min(1, t("validation.clientSecretRequired")),
  }),
);

const { handleSubmit, isSubmitting, setFieldValue } = useForm({
  validationSchema,
  initialValues: {
    name: "",
    description: "",
    gitUrl: "",
    organizationId: "",
    environmentName: "",
    environmentUrl: "",
    clientId: "",
    clientSecret: "",
  },
});

watch(organizations, (orgs) => {
  if (orgs.length > 0) {
    setFieldValue("organizationId", orgs[0].id);
  }
});

const onSubmit = handleSubmit(async (values) => {
  try {
    const { error: apiError } = await api.POST("/organizations/{orgId}/shops", {
      params: { path: { orgId: values.organizationId } },
      body: {
        name: values.name,
        description: values.description ?? undefined,
        gitUrl: values.gitUrl || undefined,
        environmentName: values.environmentName,
        environmentUrl: values.environmentUrl.replace(/\/+$/, ""),
        clientId: values.clientId,
        clientSecret: values.clientSecret,
      },
    });

    if (apiError) {
      error(apiError.message);
      return;
    }

    if (route.query.redirect) {
      router.push(route.query.redirect as string);
    } else {
      router.push({ name: "account.shop.list" });
    }
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
});

function openPluginModal() {
  showPluginModal.value = true;
  pluginBase64.value = "";
  pluginError.value = "";
}

function closePluginModal() {
  showPluginModal.value = false;
  pluginBase64.value = "";
  pluginError.value = "";
}

function processPluginData() {
  try {
    pluginError.value = "";

    if (!pluginBase64.value.trim()) {
      pluginError.value = t("environment.base64Error");
      return;
    }

    const decodedString = window.atob(pluginBase64.value.trim());
    const data = JSON.parse(decodedString);

    if (!data.url || !data.clientId || !data.clientSecret) {
      pluginError.value = t("environment.base64InvalidData");
      return;
    }

    setFieldValue("environmentUrl", data.url);
    setFieldValue("clientId", data.clientId);
    setFieldValue("clientSecret", data.clientSecret);

    closePluginModal();
  } catch (_e) {
    pluginError.value = t("environment.base64InvalidFormat");
  }
}

onMounted(() => {
  loadOrganizations();
});
</script>
