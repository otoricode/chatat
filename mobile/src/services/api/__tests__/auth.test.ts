// @ts-nocheck
jest.mock('../client', () => ({
  __esModule: true,
  default: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}));

import apiClient from '../client';
import { authApi } from '../auth';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('authApi', () => {
  it('sendOTP calls post', async () => {
    mock.post.mockResolvedValue({ data: { sessionId: 's1' } });
    await authApi.sendOTP('+628123');
    expect(mock.post).toHaveBeenCalledWith('/auth/otp/send', { phone: '+628123' });
  });

  it('verifyOTP calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await authApi.verifyOTP('+628123', '1234');
    expect(mock.post).toHaveBeenCalledWith('/auth/otp/verify', { phone: '+628123', code: '1234' });
  });

  it('initReverseOTP calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await authApi.initReverseOTP('+628123');
    expect(mock.post).toHaveBeenCalledWith('/auth/reverse-otp/init', { phone: '+628123' });
  });

  it('checkReverseOTP calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await authApi.checkReverseOTP('session1');
    expect(mock.post).toHaveBeenCalledWith('/auth/reverse-otp/check', { sessionId: 'session1' });
  });

  it('refreshToken calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await authApi.refreshToken('refresh-tok');
    expect(mock.post).toHaveBeenCalledWith('/auth/refresh', { refreshToken: 'refresh-tok' });
  });

  it('logout calls post', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await authApi.logout();
    expect(mock.post).toHaveBeenCalledWith('/auth/logout');
  });
});
