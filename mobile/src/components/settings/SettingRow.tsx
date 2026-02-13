// SettingRow — single row in settings section
import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Switch } from 'react-native';
import { colors, fontFamily, fontSize, spacing } from '@/theme';

type SettingRowProps = {
  icon?: string;
  label: string;
  value?: string;
  onPress?: () => void;
  type?: 'navigate' | 'switch';
  switchValue?: boolean;
  onToggle?: (value: boolean) => void;
  danger?: boolean;
  showDivider?: boolean;
};

export function SettingRow({
  icon,
  label,
  value,
  onPress,
  type = 'navigate',
  switchValue,
  onToggle,
  danger = false,
  showDivider = true,
}: SettingRowProps) {
  const content = (
    <View style={styles.container}>
      <View style={styles.row}>
        {icon ? <Text style={styles.icon}>{icon}</Text> : null}
        <View style={styles.labelContainer}>
          <Text style={[styles.label, danger && styles.dangerLabel]}>
            {label}
          </Text>
        </View>
        {type === 'switch' && onToggle ? (
          <Switch
            value={switchValue}
            onValueChange={onToggle}
            trackColor={{ false: colors.border, true: colors.green }}
            thumbColor={colors.white}
          />
        ) : (
          <View style={styles.right}>
            {value ? <Text style={styles.value}>{value}</Text> : null}
            {onPress ? <Text style={styles.chevron}>›</Text> : null}
          </View>
        )}
      </View>
      {showDivider && <View style={styles.divider} />}
    </View>
  );

  if (type === 'switch') {
    return content;
  }

  return (
    <TouchableOpacity onPress={onPress} activeOpacity={0.6}>
      {content}
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: spacing.lg,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.md,
    minHeight: 48,
  },
  icon: {
    fontSize: 20,
    marginRight: spacing.md,
    width: 28,
    textAlign: 'center',
  },
  labelContainer: {
    flex: 1,
  },
  label: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textPrimary,
  },
  dangerLabel: {
    color: colors.red,
  },
  right: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
  },
  value: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.md,
    color: colors.textMuted,
  },
  chevron: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.xl,
    color: colors.textMuted,
  },
  divider: {
    height: StyleSheet.hairlineWidth,
    backgroundColor: colors.border,
    marginLeft: 44,
  },
});
