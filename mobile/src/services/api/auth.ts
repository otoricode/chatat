// Auth API service
import apiClient from './client';

type OTPSendResponse = {
  sessionId: string;
};

type OTPVerifyResponse = {
  accessToken: string;
  refreshToken: string;
  user: {
    id: string;
    name: string;
    phone: string;
    avatar: string;
    status: string;
  };
  isNewUser: boolean;
};

type ReverseOTPInitResponse = {
  sessionId: string;
  waNumber: string;
  code: string;
};

type ReverseOTPCheckResponse = {
  verified: boolean;
  accessToken?: string;
  refreshToken?: string;
  user?: {
    id: string;
    name: string;
    phone: string;
    avatar: string;
    status: string;
  };
  isNewUser?: boolean;
};

type RefreshResponse = {
  accessToken: string;
  refreshToken: string;
};

export const authApi = {
  sendOTP: (phone: string) =>
    apiClient.post<OTPSendResponse>('/auth/otp/send', { phone }),

  verifyOTP: (phone: string, code: string) =>
    apiClient.post<OTPVerifyResponse>('/auth/otp/verify', { phone, code }),

  initReverseOTP: (phone: string) =>
    apiClient.post<ReverseOTPInitResponse>('/auth/reverse-otp/init', { phone }),

  checkReverseOTP: (sessionId: string) =>
    apiClient.post<ReverseOTPCheckResponse>('/auth/reverse-otp/check', {
      sessionId,
    }),

  refreshToken: (token: string) =>
    apiClient.post<RefreshResponse>('/auth/refresh', { refreshToken: token }),

  logout: () => apiClient.post('/auth/logout'),
};
