// Notification store — manages in-app notification toast state
import { create } from 'zustand';

type NotificationPayload = {
  title: string;
  body: string;
  data?: Record<string, string>;
};

type NotificationState = {
  visible: boolean;
  title: string;
  body: string;
  data: Record<string, string>;

  show: (payload: NotificationPayload) => void;
  hide: () => void;
};

export const useNotificationStore = create<NotificationState>()((set) => ({
  visible: false,
  title: '',
  body: '',
  data: {},

  show: (payload) =>
    set({
      visible: true,
      title: payload.title,
      body: payload.body,
      data: payload.data ?? {},
    }),

  hide: () =>
    set({
      visible: false,
    }),
}));
