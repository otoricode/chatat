import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { getLocales } from 'expo-localization';

import id from './id.json';
import en from './en.json';
import ar from './ar.json';

const LANGUAGE_KEY = 'user_language';

export type SupportedLanguage = 'id' | 'en' | 'ar';

const supportedLanguages: SupportedLanguage[] = ['id', 'en', 'ar'];

const resources = {
  id: { translation: id },
  en: { translation: en },
  ar: { translation: ar },
};

i18n.use(initReactI18next).init({
  resources,
  lng: 'id',
  fallbackLng: 'id',
  interpolation: {
    escapeValue: false,
  },
});

/**
 * Initialize language from saved preference or device locale.
 * Call this on app startup.
 */
export async function initLanguage(): Promise<void> {
  try {
    const saved = await AsyncStorage.getItem(LANGUAGE_KEY);
    if (saved && supportedLanguages.includes(saved as SupportedLanguage)) {
      await i18n.changeLanguage(saved);
      return;
    }

    // Detect device language
    const locales = getLocales();
    const deviceLang = locales[0]?.languageCode;
    if (deviceLang && supportedLanguages.includes(deviceLang as SupportedLanguage)) {
      await i18n.changeLanguage(deviceLang);
      await AsyncStorage.setItem(LANGUAGE_KEY, deviceLang);
    }
  } catch {
    // Fallback to default (id) silently
  }
}

/**
 * Change the app language and persist the choice.
 */
export async function setLanguage(lang: SupportedLanguage): Promise<void> {
  await i18n.changeLanguage(lang);
  await AsyncStorage.setItem(LANGUAGE_KEY, lang);
}

/**
 * Get the current language.
 */
export function getCurrentLanguage(): SupportedLanguage {
  return (i18n.language || 'id') as SupportedLanguage;
}

export default i18n;
