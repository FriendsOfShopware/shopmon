<template>
  <header-container :title="$t('onboarding.welcome')" />
  <div class="onboarding">
    <Panel>
        <div class="onboarding-hero">
          <icon-fa6-solid:building class="onboarding-icon" />
          <h2 class="onboarding-title">{{ $t("onboarding.title") }}</h2>
          <p class="onboarding-description">
            {{ $t("onboarding.description") }}
          </p>
        </div>

        <div class="onboarding-options">
          <div class="onboarding-option">
            <div class="onboarding-option-icon">
              <icon-fa6-solid:plus />
            </div>
            <div class="onboarding-option-content">
              <h3>{{ $t("onboarding.createTitle") }}</h3>
              <p>{{ $t("onboarding.createDescription") }}</p>

              <vee-form
                v-slot="{ errors, isSubmitting }"
                :validation-schema="schema"
                class="onboarding-form"
                @submit="onCreateOrganization"
              >
                <InputField
                  name="name"
                  :label="$t('common.name')"
                  :placeholder="$t('onboarding.orgNamePlaceholder')"
                  autocomplete="organization"
                  :error="errors.name"
                />

                <UiButton :disabled="isSubmitting" type="submit" variant="primary">
                  <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
                  <icon-line-md:loading-twotone-loop v-else class="icon" />
                  {{ $t("onboarding.createButton") }}
                </UiButton>
              </vee-form>
            </div>
          </div>

          <div class="onboarding-divider">
            <span>{{ $t("common.or") }}</span>
          </div>

          <div class="onboarding-option">
            <div class="onboarding-option-icon">
              <icon-fa6-solid:envelope />
            </div>
            <div class="onboarding-option-content">
              <h3>{{ $t("onboarding.inviteTitle") }}</h3>
              <p>{{ $t("onboarding.inviteDescription") }}</p>
            </div>
          </div>
        </div>
    </Panel>
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { fetchSession } from "@/composables/useSession";
import { api } from "@/helpers/api";

import { Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import * as Yup from "yup";

const { t } = useI18n();
const { error } = useAlert();
const router = useRouter();

const schema = Yup.object().shape({
  name: Yup.string().required(t("validation.orgNameRequired")),
});

async function onCreateOrganization(values: Record<string, unknown>) {
  const typedValues = values as Yup.InferType<typeof schema>;
  try {
    const { error: respError } = await api.POST("/auth/organizations", {
      body: {
        name: typedValues.name,
      },
    });

    if (respError) {
      error((respError as { message?: string }).message ?? "Failed to create organization");
      return;
    }
    await fetchSession();
    await router.push({ name: "account.dashboard" });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
}
</script>

<style scoped>
.onboarding {
  width: 100%;
}

.onboarding-hero {
  text-align: center;
  padding-bottom: 2rem;
  margin-bottom: 2rem;
  border-bottom: 1px solid var(--panel-border-color);
}

.onboarding-icon {
  width: 3rem;
  height: 3rem;
  color: var(--primary-color);
  margin-bottom: 1rem;
}

.onboarding-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.onboarding-description {
  color: var(--text-color-muted);
  max-width: 28rem;
  margin: 0 auto;
  line-height: 1.6;
}

.onboarding-options {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.onboarding-option {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.onboarding-option-icon {
  flex-shrink: 0;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.5rem;
  background-color: var(--primary-color);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;

  svg {
    width: 1rem;
    height: 1rem;
  }
}

.onboarding-option-content {
  flex: 1;

  h3 {
    font-size: 1.1rem;
    font-weight: 600;
    margin-bottom: 0.25rem;
  }

  p {
    color: var(--text-color-muted);
    font-size: 0.9rem;
    line-height: 1.5;
    margin-bottom: 0;
  }
}

.onboarding-form {
  margin-top: 1rem;
  display: flex;
  gap: 0.75rem;
  align-items: flex-end;

  .form-group {
    flex: 1;
    margin-bottom: 0;
  }

  .ui-button {
    flex-shrink: 0;
    height: fit-content;
  }
}

.onboarding-divider {
  display: flex;
  align-items: center;
  gap: 1rem;
  color: var(--text-color-muted);
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;

  &::before,
  &::after {
    content: "";
    flex: 1;
    height: 1px;
    background-color: var(--panel-border-color);
  }
}
</style>
