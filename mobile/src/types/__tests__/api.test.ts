import { isApiError, parseError } from '../api';
import type { ApiError } from '../api';

describe('api types', () => {
  describe('isApiError', () => {
    it('returns true for valid ApiError object', () => {
      const err: ApiError = { code: 'BAD_REQUEST', message: 'bad' };
      expect(isApiError(err)).toBe(true);
    });

    it('returns false for null', () => {
      expect(isApiError(null)).toBe(false);
    });

    it('returns false for undefined', () => {
      expect(isApiError(undefined)).toBe(false);
    });

    it('returns false for string', () => {
      expect(isApiError('error')).toBe(false);
    });

    it('returns false for object missing code', () => {
      expect(isApiError({ message: 'msg' })).toBe(false);
    });

    it('returns false for object missing message', () => {
      expect(isApiError({ code: 'BAD_REQUEST' })).toBe(false);
    });
  });

  describe('parseError', () => {
    it('returns ApiError if already valid', () => {
      const err: ApiError = { code: 'NOT_FOUND', message: 'not found' };
      expect(parseError(err)).toBe(err);
    });

    it('returns NETWORK_ERROR for Network errors', () => {
      const err = new Error('Network request failed');
      const result = parseError(err);
      expect(result.code).toBe('NETWORK_ERROR');
      expect(result.message).toBe('No internet connection');
    });

    it('returns INTERNAL_ERROR for other errors', () => {
      const err = new Error('Something went wrong');
      const result = parseError(err);
      expect(result.code).toBe('INTERNAL_ERROR');
      expect(result.message).toBe('Something went wrong');
    });

    it('returns INTERNAL_ERROR for unknown values', () => {
      const result = parseError(42);
      expect(result.code).toBe('INTERNAL_ERROR');
      expect(result.message).toBe('An unknown error occurred');
    });
  });
});
