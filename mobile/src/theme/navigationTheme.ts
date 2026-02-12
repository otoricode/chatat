// React Navigation theme for Chatat (dark)

import type { Theme } from '@react-navigation/native';
import { colors } from './colors';

export const navigationTheme: Theme = {
  dark: true,
  colors: {
    primary: colors.green,
    background: colors.background,
    card: colors.surface,
    text: colors.textPrimary,
    border: colors.border,
    notification: colors.red,
  },
  fonts: {
    regular: { fontFamily: 'PlusJakartaSans', fontWeight: '400' },
    medium: { fontFamily: 'PlusJakartaSans-Medium', fontWeight: '500' },
    bold: { fontFamily: 'PlusJakartaSans-Bold', fontWeight: '700' },
    heavy: { fontFamily: 'PlusJakartaSans-Bold', fontWeight: '900' },
  },
};
