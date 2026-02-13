// App-wide constants

export const APP_NAME = 'Chatat';

export const API_BASE_URL = 'http://localhost:8080/api/v1';
export const WS_BASE_URL = 'ws://localhost:8080/ws';

export const OTP_LENGTH = 6;
export const OTP_EXPIRY_SECONDS = 120;

export const MAX_GROUP_MEMBERS = 256;
export const MAX_MESSAGE_LENGTH = 4096;
export const MAX_GROUP_NAME_LENGTH = 100;

export const SUPPORTED_LOCALES = ['id', 'en', 'ar'] as const;
export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number];
export const DEFAULT_LOCALE: SupportedLocale = 'id';
