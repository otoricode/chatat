jest.mock('@react-native-community/netinfo', () => {
  let listener: ((state: any) => void) | null = null;
  return {
    addEventListener: jest.fn((fn: any) => {
      listener = fn;
      return jest.fn(); // unsubscribe
    }),
    _triggerChange: (state: any) => {
      if (listener) listener(state);
    },
    fetch: jest.fn(),
  };
});

jest.mock('@/services/MessageQueue', () => ({
  messageQueue: {
    flushPending: jest.fn(),
  },
}));

import { useNetworkStore } from '../networkStore';
import NetInfo from '@react-native-community/netinfo';
import { messageQueue } from '@/services/MessageQueue';

const mockNetInfo = NetInfo as any;
const mockFlush = messageQueue.flushPending as jest.Mock;

beforeEach(() => {
  useNetworkStore.setState({
    isConnected: true,
    isInternetReachable: null,
    type: 'unknown',
  });
  jest.clearAllMocks();
});

describe('networkStore', () => {
  it('starts connected', () => {
    const s = useNetworkStore.getState();
    expect(s.isConnected).toBe(true);
    expect(s.type).toBe('unknown');
  });

  it('startListening subscribes to NetInfo', () => {
    const unsub = useNetworkStore.getState().startListening();

    expect(mockNetInfo.addEventListener).toHaveBeenCalled();
    expect(typeof unsub).toBe('function');
  });

  it('updates state when network changes', () => {
    useNetworkStore.getState().startListening();

    mockNetInfo._triggerChange({
      isConnected: false,
      isInternetReachable: false,
      type: 'none',
    });

    const s = useNetworkStore.getState();
    expect(s.isConnected).toBe(false);
    expect(s.isInternetReachable).toBe(false);
    expect(s.type).toBe('none');
  });

  it('flushes pending when transitioning online', () => {
    useNetworkStore.setState({ isConnected: false });
    useNetworkStore.getState().startListening();

    mockNetInfo._triggerChange({
      isConnected: true,
      isInternetReachable: true,
      type: 'wifi',
    });

    expect(mockFlush).toHaveBeenCalled();
  });

  it('does not flush when staying online', () => {
    useNetworkStore.setState({ isConnected: true });
    useNetworkStore.getState().startListening();

    mockNetInfo._triggerChange({
      isConnected: true,
      isInternetReachable: true,
      type: 'wifi',
    });

    expect(mockFlush).not.toHaveBeenCalled();
  });

  it('handles null isConnected as false', () => {
    useNetworkStore.getState().startListening();

    mockNetInfo._triggerChange({
      isConnected: null,
      isInternetReachable: null,
      type: 'unknown',
    });

    expect(useNetworkStore.getState().isConnected).toBe(false);
  });
});
