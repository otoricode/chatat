// @ts-nocheck
import {
  APP_NAME,
  API_BASE_URL,
  WS_BASE_URL,
  OTP_LENGTH,
  OTP_EXPIRY_SECONDS,
  MAX_GROUP_MEMBERS,
  MAX_MESSAGE_LENGTH,
  MAX_GROUP_NAME_LENGTH,
  SUPPORTED_LOCALES,
  DEFAULT_LOCALE,
} from '../constants';

describe('constants', () => {
  it('APP_NAME is Chatat', () => {
    expect(APP_NAME).toBe('Chatat');
  });

  it('API_BASE_URL is defined', () => {
    expect(API_BASE_URL).toContain('http');
  });

  it('WS_BASE_URL is defined', () => {
    expect(WS_BASE_URL).toContain('ws');
  });

  it('OTP_LENGTH is 6', () => {
    expect(OTP_LENGTH).toBe(6);
  });

  it('OTP_EXPIRY_SECONDS is 120', () => {
    expect(OTP_EXPIRY_SECONDS).toBe(120);
  });

  it('MAX_GROUP_MEMBERS is 256', () => {
    expect(MAX_GROUP_MEMBERS).toBe(256);
  });

  it('MAX_MESSAGE_LENGTH is 4096', () => {
    expect(MAX_MESSAGE_LENGTH).toBe(4096);
  });

  it('MAX_GROUP_NAME_LENGTH is 100', () => {
    expect(MAX_GROUP_NAME_LENGTH).toBe(100);
  });

  it('SUPPORTED_LOCALES includes id, en, ar', () => {
    expect(SUPPORTED_LOCALES).toContain('id');
    expect(SUPPORTED_LOCALES).toContain('en');
    expect(SUPPORTED_LOCALES).toContain('ar');
    expect(SUPPORTED_LOCALES).toHaveLength(3);
  });

  it('DEFAULT_LOCALE is id', () => {
    expect(DEFAULT_LOCALE).toBe('id');
  });
});
