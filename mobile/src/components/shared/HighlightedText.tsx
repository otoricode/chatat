// HighlightedText — renders text with <mark> tags as highlighted spans
import React from 'react';
import { Text, type TextStyle } from 'react-native';
import { colors, fontFamily, fontSize } from '@/theme';

type HighlightedTextProps = {
  text: string;
  style?: TextStyle;
};

export function HighlightedText({ text, style }: HighlightedTextProps) {
  const parts = text.split(/(<mark>.*?<\/mark>)/g);

  return (
    <Text style={[defaultStyle, style]} numberOfLines={2}>
      {parts.map((part, i) => {
        if (part.startsWith('<mark>')) {
          const content = part.replace(/<\/?mark>/g, '');
          return (
            <Text key={i} style={highlightStyle}>
              {content}
            </Text>
          );
        }
        return <Text key={i}>{part}</Text>;
      })}
    </Text>
  );
}

const defaultStyle: TextStyle = {
  fontFamily: fontFamily.ui,
  fontSize: fontSize.sm,
  color: colors.textMuted,
  lineHeight: 18,
};

const highlightStyle: TextStyle = {
  color: colors.green,
  fontFamily: fontFamily.uiSemiBold,
};
