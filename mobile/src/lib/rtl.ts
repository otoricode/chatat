import { I18nManager } from 'react-native';

/**
 * Check if the current layout direction is RTL.
 */
export function isRTL(): boolean {
  return I18nManager.isRTL;
}

/**
 * Apply RTL settings based on language.
 * Returns true if a restart is needed (RTL state changed).
 */
export function applyRTL(language: string): boolean {
  const shouldBeRTL = language === 'ar';

  if (I18nManager.isRTL !== shouldBeRTL) {
    I18nManager.allowRTL(shouldBeRTL);
    I18nManager.forceRTL(shouldBeRTL);
    return true; // restart needed
  }

  return false;
}

/**
 * Directional margin helpers that respect RTL.
 */
export const marginStart = (value: number) =>
  I18nManager.isRTL ? { marginRight: value } : { marginLeft: value };

export const marginEnd = (value: number) =>
  I18nManager.isRTL ? { marginLeft: value } : { marginRight: value };

export const paddingStart = (value: number) =>
  I18nManager.isRTL ? { paddingRight: value } : { paddingLeft: value };

export const paddingEnd = (value: number) =>
  I18nManager.isRTL ? { paddingLeft: value } : { paddingRight: value };

/**
 * Get the flex direction for a row that respects RTL.
 */
export const rowDirection = () =>
  I18nManager.isRTL ? ('row-reverse' as const) : ('row' as const);

/**
 * Get text alignment that respects RTL.
 */
export const textAlign = () =>
  I18nManager.isRTL ? ('right' as const) : ('left' as const);
