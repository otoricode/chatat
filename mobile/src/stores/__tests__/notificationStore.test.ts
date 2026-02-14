// @ts-nocheck
import { useNotificationStore } from '../notificationStore';

beforeEach(() => {
  useNotificationStore.setState({
    visible: false,
    title: '',
    body: '',
    data: {},
  });
});

describe('notificationStore', () => {
  it('starts hidden', () => {
    const s = useNotificationStore.getState();
    expect(s.visible).toBe(false);
    expect(s.title).toBe('');
    expect(s.body).toBe('');
    expect(s.data).toEqual({});
  });

  it('show sets visible and payload', () => {
    useNotificationStore.getState().show({
      title: 'New Message',
      body: 'Hello there',
      data: { chatId: '123' },
    });

    const s = useNotificationStore.getState();
    expect(s.visible).toBe(true);
    expect(s.title).toBe('New Message');
    expect(s.body).toBe('Hello there');
    expect(s.data).toEqual({ chatId: '123' });
  });

  it('show without data defaults to empty object', () => {
    useNotificationStore.getState().show({
      title: 'Alert',
      body: 'Something happened',
    });

    expect(useNotificationStore.getState().data).toEqual({});
  });

  it('hide clears visible', () => {
    useNotificationStore.getState().show({ title: 'T', body: 'B' });
    useNotificationStore.getState().hide();

    expect(useNotificationStore.getState().visible).toBe(false);
  });
});
