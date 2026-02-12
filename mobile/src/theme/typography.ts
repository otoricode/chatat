// Typography system for Chatat
// Based on spesifikasi-chatat.md section 9.3

export const fontFamily = {
  ui: 'PlusJakartaSans',
  uiBold: 'PlusJakartaSans-Bold',
  uiSemiBold: 'PlusJakartaSans-SemiBold',
  uiMedium: 'PlusJakartaSans-Medium',
  document: 'Inter',
  documentBold: 'Inter-Bold',
  documentMedium: 'Inter-Medium',
  code: 'JetBrainsMono',
} as const;

export const fontSize = {
  xs: 11,
  sm: 13,
  md: 15,
  lg: 17,
  xl: 20,
  xxl: 24,
  h1: 28,
  h2: 24,
  h3: 20,
} as const;

export const lineHeight = {
  tight: 1.2,
  normal: 1.5,
  relaxed: 1.75,
} as const;

export type FontSizeKey = keyof typeof fontSize;
