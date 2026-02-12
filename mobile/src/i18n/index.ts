import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import id from './id.json';
import en from './en.json';
import ar from './ar.json';

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

export default i18n;
