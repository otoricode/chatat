// Time formatting utilities for chat
// Formats timestamps for chat list and message bubbles

/**
 * Format a timestamp for the chat list.
 * - Today: "HH:MM"
 * - Yesterday: "Kemarin"
 * - This week: day name (Sen, Sel, Rab, ...)
 * - Older: "DD/MM/YY"
 */
export function formatChatListTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();

  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (isToday(date, now)) {
    return formatTime(date);
  }

  if (diffDays === 1 || isYesterday(date, now)) {
    return 'Kemarin';
  }

  if (diffDays < 7) {
    return getDayName(date);
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
 * Format a date separator: "Hari Ini", "Kemarin", or "DD MMM YYYY"
 */
export function formatDateSeparator(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();

  if (isToday(date, now)) {
    return 'Hari Ini';
  }

  if (isYesterday(date, now)) {
    return 'Kemarin';
  }

  const months = [
    'Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun',
    'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des',
  ];

  return `${date.getDate()} ${months[date.getMonth()]} ${date.getFullYear()}`;
}

/**
 * Format last seen: "terakhir dilihat pukul HH:MM" or "online"
 */
export function formatLastSeen(dateStr: string, isOnline: boolean): string {
  if (isOnline) return 'online';
  const date = new Date(dateStr);
  const now = new Date();

  if (isToday(date, now)) {
    return `terakhir dilihat pukul ${formatTime(date)}`;
  }

  if (isYesterday(date, now)) {
    return `terakhir dilihat kemarin pukul ${formatTime(date)}`;
  }

  return `terakhir dilihat ${formatDate(date)} pukul ${formatTime(date)}`;
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

function getDayName(date: Date): string {
  const days = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];
  return days[date.getDay()] ?? 'Min';
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
