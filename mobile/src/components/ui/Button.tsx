// Button component — primary, secondary, danger, ghost variants
import React from 'react';
import {
  Pressable,
  Text,
  StyleSheet,
  ActivityIndicator,
  type ViewStyle,
} from 'react-native';
import { colors, fontSize, spacing, fontFamily } from '@/theme';

type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost';

type ButtonProps = {
  title: string;
  variant?: ButtonVariant;
  onPress: () => void;
  disabled?: boolean;
  loading?: boolean;
  style?: ViewStyle;
};

export function Button({
  title,
  variant = 'primary',
  onPress,
  disabled = false,
  loading = false,
  style,
}: ButtonProps) {
  const isDisabled = disabled || loading;

  return (
    <Pressable
      style={({ pressed }) => [
        styles.button,
        variantStyles[variant],
        pressed && styles.pressed,
        isDisabled && styles.disabled,
        style,
      ]}
      onPress={onPress}
      disabled={isDisabled}
    >
      {loading ? (
        <ActivityIndicator
          size="small"
          color={variant === 'primary' ? colors.background : colors.green}
        />
      ) : (
        <Text style={[styles.text, textVariantStyles[variant]]}>{title}</Text>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: {
    borderRadius: 12,
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.xl,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },
  pressed: {
    opacity: 0.8,
  },
  disabled: {
    opacity: 0.4,
  },
  text: {
    fontFamily: fontFamily.uiSemiBold,
    fontSize: fontSize.md,
  },
});

const variantStyles = StyleSheet.create({
  primary: {
    backgroundColor: colors.green,
  },
  secondary: {
    backgroundColor: colors.transparent,
    borderWidth: 1,
    borderColor: colors.green,
  },
  danger: {
    backgroundColor: colors.red,
  },
  ghost: {
    backgroundColor: colors.transparent,
  },
});

const textVariantStyles = StyleSheet.create({
  primary: {
    color: colors.background,
  },
  secondary: {
    color: colors.green,
  },
  danger: {
    color: colors.white,
  },
  ghost: {
    color: colors.green,
  },
});
