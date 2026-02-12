// Theme constants for Chatat
// Based on WhatsApp dark style with green accent

export const colors = {
  // Background
  background: '#0F1117',
  surface: '#1A1D27',
  surfaceLight: '#252836',

  // Primary
  primary: '#6EE7B7',
  primaryDark: '#34D399',

  // Text
  textPrimary: '#F9FAFB',
  textSecondary: '#9CA3AF',
  textMuted: '#6B7280',

  // Status
  success: '#6EE7B7',
  error: '#EF4444',
  warning: '#F59E0B',
  info: '#3B82F6',

  // Chat
  bubbleSent: '#065F46',
  bubbleReceived: '#252836',

  // Border
  border: '#374151',
  divider: '#1F2937',
} as const;

export const spacing = {
  xs: 4,
  sm: 8,
  md: 16,
  lg: 24,
  xl: 32,
} as const;

export const fontSize = {
  xs: 12,
  sm: 14,
  md: 16,
  lg: 18,
  xl: 20,
  xxl: 24,
  title: 32,
} as const;
