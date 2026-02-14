// @ts-nocheck
import {
  formatChatListTime,
  formatMessageTime,
  formatDateSeparator,
  formatLastSeen,
  isDifferentDay,
} from '../timeFormat';

// Mock translation function
const t = (key: string, params?: Record<string, string>) => {
  const translations: Record<string, string> = {
    'time.yesterday': 'Yesterday',
    'time.today': 'Today',
    'time.sun': 'Sun',
    'time.mon': 'Mon',
    'time.tue': 'Tue',
    'time.wed': 'Wed',
    'time.thu': 'Thu',
    'time.fri': 'Fri',
    'time.sat': 'Sat',
    'time.jan': 'Jan',
    'time.feb': 'Feb',
    'time.mar': 'Mar',
    'time.apr': 'Apr',
    'time.may': 'May',
    'time.jun': 'Jun',
    'time.jul': 'Jul',
    'time.aug': 'Aug',
    'time.sep': 'Sep',
    'time.oct': 'Oct',
    'time.nov': 'Nov',
    'time.dec': 'Dec',
    'time.lastSeenAt': `last seen at ${params?.time ?? ''}`,
    'time.lastSeenYesterdayAt': `last seen yesterday at ${params?.time ?? ''}`,
    'time.lastSeenDateAt': `last seen ${params?.date ?? ''} at ${params?.time ?? ''}`,
  };
  return translations[key] ?? key;
};

describe('formatChatListTime', () => {
  it('returns time for today', () => {
    const now = new Date();
    now.setHours(14, 30);
    const result = formatChatListTime(now.toISOString(), t as any);
    expect(result).toBe('14:30');
  });

  it('returns Yesterday for yesterday', () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    yesterday.setHours(10, 0);
    const result = formatChatListTime(yesterday.toISOString(), t as any);
    expect(result).toBe('Yesterday');
  });

  it('returns day name for this week', () => {
    const date = new Date();
    date.setDate(date.getDate() - 3);
    date.setHours(10, 0);
    const result = formatChatListTime(date.toISOString(), t as any);
    // Should be one of the day names
    expect(['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']).toContain(result);
  });

  it('returns DD/MM/YY for older dates', () => {
    const old = new Date('2024-01-15T10:00:00Z');
    const result = formatChatListTime(old.toISOString(), t as any);
    expect(result).toMatch(/^\d{2}\/\d{2}\/\d{2}$/);
  });
});

describe('formatMessageTime', () => {
  it('returns HH:MM format', () => {
    const date = new Date();
    date.setHours(9, 5, 0, 0);
    const result = formatMessageTime(date.toISOString());
    expect(result).toBe('09:05');
  });

  it('pads single digit hours', () => {
    const date = new Date();
    date.setHours(3, 7);
    const result = formatMessageTime(date.toISOString());
    expect(result).toMatch(/^0\d:\d{2}$/);
  });
});

describe('formatDateSeparator', () => {
  it('returns Today for today', () => {
    const now = new Date();
    const result = formatDateSeparator(now.toISOString(), t as any);
    expect(result).toBe('Today');
  });

  it('returns Yesterday for yesterday', () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    const result = formatDateSeparator(yesterday.toISOString(), t as any);
    expect(result).toBe('Yesterday');
  });

  it('returns DD MMM YYYY for older dates', () => {
    const old = new Date('2024-03-15T10:00:00Z');
    const result = formatDateSeparator(old.toISOString(), t as any);
    expect(result).toContain('Mar');
    expect(result).toContain('2024');
  });
});

describe('formatLastSeen', () => {
  it('returns online if user is online', () => {
    const result = formatLastSeen(new Date().toISOString(), true, t as any);
    expect(result).toBe('online');
  });

  it('returns last seen at HH:MM for today', () => {
    const now = new Date();
    now.setHours(14, 30);
    const result = formatLastSeen(now.toISOString(), false, t as any);
    expect(result).toContain('last seen at');
    expect(result).toContain('14:30');
  });

  it('returns last seen yesterday for yesterday', () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    yesterday.setHours(10, 0);
    const result = formatLastSeen(yesterday.toISOString(), false, t as any);
    expect(result).toContain('yesterday');
  });

  it('returns last seen date for older dates', () => {
    const old = new Date('2024-01-15T10:00:00Z');
    const result = formatLastSeen(old.toISOString(), false, t as any);
    expect(result).toContain('last seen');
  });
});

describe('isDifferentDay', () => {
  it('returns false for same day', () => {
    const d1 = '2024-03-15T10:00:00';
    const d2 = '2024-03-15T18:00:00';
    expect(isDifferentDay(d1, d2)).toBe(false);
  });

  it('returns true for different days', () => {
    const d1 = '2024-03-15T10:00:00Z';
    const d2 = '2024-03-16T10:00:00Z';
    expect(isDifferentDay(d1, d2)).toBe(true);
  });

  it('returns true for different months', () => {
    const d1 = '2024-03-15T10:00:00Z';
    const d2 = '2024-04-15T10:00:00Z';
    expect(isDifferentDay(d1, d2)).toBe(true);
  });

  it('returns true for different years', () => {
    const d1 = '2024-03-15T10:00:00Z';
    const d2 = '2025-03-15T10:00:00Z';
    expect(isDifferentDay(d1, d2)).toBe(true);
  });
});
