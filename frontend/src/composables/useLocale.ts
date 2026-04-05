import { useI18n } from "vue-i18n";

const LOCALE_STORAGE_KEY = "shopmon-locale";

export function useLocale() {
  const { locale } = useI18n();

  function setLocale(lang: string) {
    locale.value = lang;
    if (!import.meta.env.SSR) {
      localStorage.setItem(LOCALE_STORAGE_KEY, lang);
      document.documentElement.lang = lang;
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
