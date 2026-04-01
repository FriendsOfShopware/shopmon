import { createI18n } from "vue-i18n";
import en from "./locales/en.json";
import de from "./locales/de.json";

const LOCALE_STORAGE_KEY = "shopmon-locale";

function getInitialLocale(): string {
  const stored = localStorage.getItem(LOCALE_STORAGE_KEY);
  if (stored && ["en", "de"].includes(stored)) {
    return stored;
  }
  const browserLang = navigator.language.split("-")[0];
  return browserLang === "de" ? "de" : "en";
}

export const i18n = createI18n<[typeof en], "en" | "de">({
  legacy: false,
  locale: getInitialLocale(),
  fallbackLocale: "en",
  messages: { en, de },
});
