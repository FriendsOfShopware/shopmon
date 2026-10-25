import { requireMutation } from "@/helpers/apiError";
import { useAlert } from "@/composables/useAlert";
import { useI18n } from "vue-i18n";
import { getToken } from "@/helpers/api";
import { updateAccountMe } from "@/api/generated";

const LOCALE_STORAGE_KEY = "shopmon-locale";

export function useLocale() {
  const { locale } = useI18n();
  const { error } = useAlert();

  function setLocale(lang: string, persist = true) {
    locale.value = lang;
    if (!import.meta.env.SSR) {
      localStorage.setItem(LOCALE_STORAGE_KEY, lang);
      document.documentElement.lang = lang;

      // Persist to the server so notification emails are sent in the same
      // language. Fire-and-forget: a failure here must not block the UI switch.
      if (persist && getToken()) {
        void updateAccountMe({ body: { locale: lang } })
          .then((result) =>
            requireMutation(
              result,
              "Browser language changed, but email language could not be saved. Please select the language again to retry.",
            ),
          )
          .catch(() =>
            error(
              "Browser language changed, but email language could not be saved. Please select the language again to retry.",
            ),
          );
      }
    }
  }

  function toggleLocale() {
    const next = locale.value === "en" ? "de" : "en";
    setLocale(next);
  }

  return {
    locale,
    setLocale,
    toggleLocale,
  };
}
