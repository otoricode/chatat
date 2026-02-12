// API request/response types

export interface ApiError {
  code: ErrorCode;
  message: string;
}

export type ErrorCode =
  | 'BAD_REQUEST'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'CONFLICT'
  | 'VALIDATION_ERROR'
  | 'INTERNAL_ERROR'
  | 'NETWORK_ERROR';

export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'code' in error &&
    'message' in error
  );
}

export function parseError(error: unknown): ApiError {
  if (isApiError(error)) return error;
  if (error instanceof Error) {
    if (error.message.includes('Network')) {
      return { code: 'NETWORK_ERROR', message: 'No internet connection' };
    }
    return { code: 'INTERNAL_ERROR', message: error.message };
  }
  return { code: 'INTERNAL_ERROR', message: 'An unknown error occurred' };
}
