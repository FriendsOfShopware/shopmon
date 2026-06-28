import { useI18n } from "vue-i18n";
import { api, getToken } from "@/helpers/api";

const LOCALE_STORAGE_KEY = "shopmon-locale";

export function useLocale() {
  const { locale } = useI18n();

  function setLocale(lang: string, persist = true) {
    locale.value = lang;
    if (!import.meta.env.SSR) {
      localStorage.setItem(LOCALE_STORAGE_KEY, lang);
      document.documentElement.lang = lang;

      // Persist to the server so notification emails are sent in the same
      // language. Fire-and-forget: a failure here must not block the UI switch.
      if (persist && getToken()) {
        void api.PATCH("/account/me", { body: { locale: lang } });
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
