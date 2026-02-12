// Color palette for Chatat
// Based on spesifikasi-chatat.md section 9.2

export const colors = {
  // Background
  background: '#0F1117',
  surface: '#1A1D27',
  surface2: '#222637',

  // Border
  border: '#2E3348',

  // Text
  textPrimary: '#E8EAF0',
  textMuted: '#6B7280',

  // Accent
  green: '#6EE7B7',
  purple: '#818CF8',
  blue: '#60A5FA',
  red: '#F87171',
  yellow: '#FBBF24',

  // Chat bubbles
  bubbleSelf: '#1B3A2D',
  bubbleOther: '#222637',
  bubbleSelfText: '#E8EAF0',
  bubbleOtherText: '#E8EAF0',

  // Status
  online: '#6EE7B7',
  offline: '#6B7280',

  // Misc
  overlay: 'rgba(0, 0, 0, 0.5)',
  inputBackground: '#1A1D27',
  tabBarBackground: '#1A1D27',
  headerBackground: '#1A1D27',

  // Transparent
  transparent: 'transparent',
  white: '#FFFFFF',
  black: '#000000',
} as const;

export type ColorKey = keyof typeof colors;
