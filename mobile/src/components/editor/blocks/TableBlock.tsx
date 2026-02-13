// TableBlock — simple table editor with rows and columns
import React, { useCallback } from 'react';
import { View, TextInput, Text, ScrollView, Pressable, StyleSheet } from 'react-native';
import { colors, fontFamily, fontSize, spacing } from '@/theme';
import type { BlockProps } from '../types';

interface TableColumn {
  name: string;
  type: 'text' | 'number' | 'date' | 'checkbox';
}

export const TableBlock = React.memo(function TableBlock({
  block,
  readOnly,
  onChange,
  onFocus,
}: BlockProps) {
  const columns: TableColumn[] = (() => {
    try {
      if (Array.isArray(block.columns)) return block.columns as TableColumn[];
      return [{ name: 'Kolom 1', type: 'text' as const }];
    } catch {
      return [{ name: 'Kolom 1', type: 'text' as const }];
    }
  })();

  const rows: string[][] = (() => {
    try {
      if (Array.isArray(block.rows)) return block.rows as string[][];
      return [columns.map(() => '')];
    } catch {
      return [columns.map(() => '')];
    }
  })();

  const updateCell = useCallback(
    (rowIndex: number, colIndex: number, value: string) => {
      const newRows = rows.map((row, ri) =>
        ri === rowIndex ? row.map((cell, ci) => (ci === colIndex ? value : cell)) : [...row],
      );
      onChange({ rows: newRows });
    },
    [rows, onChange],
  );

  const updateHeader = useCallback(
    (colIndex: number, name: string) => {
      const newCols = columns.map((col, ci) =>
        ci === colIndex ? { ...col, name } : col,
      );
      onChange({ columns: newCols });
    },
    [columns, onChange],
  );

  const addRow = useCallback(() => {
    const newRow = columns.map(() => '');
    onChange({ rows: [...rows, newRow] });
  }, [columns, rows, onChange]);

  const addColumn = useCallback(() => {
    const newCols = [...columns, { name: `Kolom ${columns.length + 1}`, type: 'text' as const }];
    const newRows = rows.map((row) => [...row, '']);
    onChange({ columns: newCols, rows: newRows });
  }, [columns, rows, onChange]);

  return (
    <View style={styles.container} onTouchStart={onFocus}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false}>
        <View>
          {/* Header */}
          <View style={styles.headerRow}>
            {columns.map((col, ci) => (
              <TextInput
                key={`h-${ci}`}
                style={styles.headerCell}
                value={col.name}
                onChangeText={(text) => updateHeader(ci, text)}
                editable={!readOnly}
              />
            ))}
            {!readOnly && (
              <Pressable onPress={addColumn} style={styles.addColBtn}>
                <Text style={styles.addBtnText}>+</Text>
              </Pressable>
            )}
          </View>
          {/* Rows */}
          {rows.map((row, ri) => (
            <View key={`r-${ri}`} style={styles.dataRow}>
              {row.map((cell, ci) => (
                <TextInput
                  key={`c-${ri}-${ci}`}
                  style={styles.dataCell}
                  value={cell}
                  onChangeText={(text) => updateCell(ri, ci, text)}
                  editable={!readOnly}
                  placeholder="..."
                  placeholderTextColor={colors.textMuted}
                />
              ))}
            </View>
          ))}
          {/* Add row */}
          {!readOnly && (
            <Pressable onPress={addRow} style={styles.addRowBtn}>
              <Text style={styles.addRowText}>+ Tambah Baris</Text>
            </Pressable>
          )}
        </View>
      </ScrollView>
    </View>
  );
});

const CELL_WIDTH = 120;

const styles = StyleSheet.create({
  container: {
    marginHorizontal: spacing.lg,
    marginVertical: spacing.xs,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.border,
    overflow: 'hidden',
  },
  headerRow: {
    flexDirection: 'row',
    backgroundColor: colors.surface2,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  headerCell: {
    width: CELL_WIDTH,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.sm,
    fontFamily: fontFamily.documentMedium,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    borderRightWidth: 1,
    borderRightColor: colors.border,
  },
  addColBtn: {
    width: 36,
    justifyContent: 'center',
    alignItems: 'center',
  },
  addBtnText: {
    fontSize: fontSize.lg,
    color: colors.textMuted,
  },
  dataRow: {
    flexDirection: 'row',
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  dataCell: {
    width: CELL_WIDTH,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.sm,
    fontFamily: fontFamily.document,
    fontSize: fontSize.sm,
    color: colors.textPrimary,
    borderRightWidth: 1,
    borderRightColor: colors.border,
  },
  addRowBtn: {
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.md,
  },
  addRowText: {
    fontFamily: fontFamily.ui,
    fontSize: fontSize.sm,
    color: colors.textMuted,
  },
});
