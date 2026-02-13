// Time formatting utilities for chat
// Formats timestamps for chat list and message bubbles
import type { TFunction } from 'i18next';

const DAY_KEYS = ['time.sun', 'time.mon', 'time.tue', 'time.wed', 'time.thu', 'time.fri', 'time.sat'] as const;
const MONTH_KEYS = ['time.jan', 'time.feb', 'time.mar', 'time.apr', 'time.may', 'time.jun', 'time.jul', 'time.aug', 'time.sep', 'time.oct', 'time.nov', 'time.dec'] as const;

/**
 * Format a timestamp for the chat list.
 * - Today: "HH:MM"
 * - Yesterday: translated
 * - This week: translated day name
 * - Older: "DD/MM/YY"
 */
export function formatChatListTime(dateStr: string, t: TFunction): string {
  const date = new Date(dateStr);
  const now = new Date();

  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (isToday(date, now)) {
    return formatTime(date);
  }

  if (diffDays === 1 || isYesterday(date, now)) {
    return t('time.yesterday');
  }

  if (diffDays < 7) {
    return t(DAY_KEYS[date.getDay()] ?? 'time.sun');
  }

  return formatDate(date);
}

/**
 * Format a timestamp for message bubbles: "HH:MM"
 */
export function formatMessageTime(dateStr: string): string {
  const date = new Date(dateStr);
  return formatTime(date);
}

/**
 * Format a date separator: translated "Today", "Yesterday", or "DD MMM YYYY"
 */
export function formatDateSeparator(dateStr: string, t: TFunction): string {
  const date = new Date(dateStr);
  const now = new Date();

  if (isToday(date, now)) {
    return t('time.today');
  }

  if (isYesterday(date, now)) {
    return t('time.yesterday');
  }

  const monthKey = MONTH_KEYS[date.getMonth()] ?? 'time.jan';

  return `${date.getDate()} ${t(monthKey)} ${date.getFullYear()}`;
}

/**
 * Format last seen with translated strings
 */
export function formatLastSeen(dateStr: string, isOnline: boolean, t: TFunction): string {
  if (isOnline) return 'online';
  const date = new Date(dateStr);
  const now = new Date();
  const time = formatTime(date);

  if (isToday(date, now)) {
    return t('time.lastSeenAt', { time });
  }

  if (isYesterday(date, now)) {
    return t('time.lastSeenYesterdayAt', { time });
  }

  return t('time.lastSeenDateAt', { date: formatDate(date), time });
}

function formatTime(date: Date): string {
  const h = String(date.getHours()).padStart(2, '0');
  const m = String(date.getMinutes()).padStart(2, '0');
  return `${h}:${m}`;
}

function formatDate(date: Date): string {
  const d = String(date.getDate()).padStart(2, '0');
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const y = String(date.getFullYear()).slice(2);
  return `${d}/${m}/${y}`;
}

function isToday(date: Date, now: Date): boolean {
  return (
    date.getDate() === now.getDate() &&
    date.getMonth() === now.getMonth() &&
    date.getFullYear() === now.getFullYear()
  );
}

function isYesterday(date: Date, now: Date): boolean {
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  return (
    date.getDate() === yesterday.getDate() &&
    date.getMonth() === yesterday.getMonth() &&
    date.getFullYear() === yesterday.getFullYear()
  );
}

/**
 * Check if two date strings are on different days (for date separators).
 */
export function isDifferentDay(dateStr1: string, dateStr2: string): boolean {
  const d1 = new Date(dateStr1);
  const d2 = new Date(dateStr2);
  return (
    d1.getDate() !== d2.getDate() ||
    d1.getMonth() !== d2.getMonth() ||
    d1.getFullYear() !== d2.getFullYear()
  );
}
