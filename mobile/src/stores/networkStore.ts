// Network store — tracks online/offline status
import { create } from 'zustand';
import NetInfo, { NetInfoState } from '@react-native-community/netinfo';
import { messageQueue } from '@/services/MessageQueue';

type NetworkState = {
  isConnected: boolean;
  isInternetReachable: boolean | null;
  type: string;
  /** Start listening to network changes */
  startListening: () => () => void;
};

export const useNetworkStore = create<NetworkState>((set, get) => ({
  isConnected: true,
  isInternetReachable: null,
  type: 'unknown',

  startListening: () => {
    const unsubscribe = NetInfo.addEventListener((state: NetInfoState) => {
      const wasConnected = get().isConnected;
      const isNowConnected = state.isConnected ?? false;

      set({
        isConnected: isNowConnected,
        isInternetReachable: state.isInternetReachable,
        type: state.type,
      });

      // When transitioning from offline to online, flush pending messages
      if (!wasConnected && isNowConnected) {
        messageQueue.flushPending();
      }
    });

    return unsubscribe;
  },
}));
