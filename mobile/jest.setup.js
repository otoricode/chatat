// Mock React Native modules for unit testing

// AsyncStorage mock
jest.mock('@react-native-async-storage/async-storage', () => {
  let store = {};
  return {
    __esModule: true,
    default: {
      getItem: jest.fn((key) => Promise.resolve(store[key] || null)),
      setItem: jest.fn((key, value) => {
        store[key] = value;
        return Promise.resolve();
      }),
      removeItem: jest.fn((key) => {
        delete store[key];
        return Promise.resolve();
      }),
      clear: jest.fn(() => {
        store = {};
        return Promise.resolve();
      }),
      multiGet: jest.fn((keys) =>
        Promise.resolve(keys.map((k) => [k, store[k] || null]))
      ),
      multiSet: jest.fn((pairs) => {
        pairs.forEach(([k, v]) => {
          store[k] = v;
        });
        return Promise.resolve();
      }),
    },
  };
});

// react-native mock
jest.mock('react-native', () => ({
  I18nManager: {
    isRTL: false,
    allowRTL: jest.fn(),
    forceRTL: jest.fn(),
  },
  Platform: { OS: 'ios', select: jest.fn((obj) => obj.ios) },
  NativeModules: {},
  StyleSheet: { create: (s) => s },
  Alert: { alert: jest.fn() },
}));

// expo-sqlite mock
jest.mock('expo-sqlite', () => ({
  openDatabaseSync: jest.fn(() => ({
    execSync: jest.fn(),
    getAllSync: jest.fn(() => []),
    getFirstSync: jest.fn(() => null),
    runSync: jest.fn(() => ({ changes: 0 })),
  })),
}));

// expo-notifications mock
jest.mock('expo-notifications', () => ({
  getPermissionsAsync: jest.fn(() =>
    Promise.resolve({ status: 'granted' })
  ),
  requestPermissionsAsync: jest.fn(() =>
    Promise.resolve({ status: 'granted' })
  ),
  getExpoPushTokenAsync: jest.fn(() =>
    Promise.resolve({ data: 'mock-token' })
  ),
  setNotificationHandler: jest.fn(),
  addNotificationReceivedListener: jest.fn(() => ({ remove: jest.fn() })),
  addNotificationResponseReceivedListener: jest.fn(() => ({
    remove: jest.fn(),
  })),
  scheduleNotificationAsync: jest.fn(),
}));

// @react-native-community/netinfo mock
jest.mock('@react-native-community/netinfo', () => ({
  addEventListener: jest.fn(() => jest.fn()),
  fetch: jest.fn(() =>
    Promise.resolve({ isConnected: true, isInternetReachable: true })
  ),
}));

// react-native-mmkv mock
jest.mock('react-native-mmkv', () => {
  let store = {};
  return {
    MMKV: jest.fn().mockImplementation(() => ({
      getString: jest.fn((key) => store[key]),
      set: jest.fn((key, value) => {
        store[key] = value;
      }),
      delete: jest.fn((key) => {
        delete store[key];
      }),
      contains: jest.fn((key) => key in store),
      clearAll: jest.fn(() => {
        store = {};
      }),
    })),
  };
});

// expo-clipboard mock
jest.mock('expo-clipboard', () => ({
  setStringAsync: jest.fn(() => Promise.resolve(true)),
  getStringAsync: jest.fn(() => Promise.resolve('')),
}));

// expo-image-picker mock
jest.mock('expo-image-picker', () => ({
  launchImageLibraryAsync: jest.fn(),
  launchCameraAsync: jest.fn(),
  requestMediaLibraryPermissionsAsync: jest.fn(() =>
    Promise.resolve({ status: 'granted' })
  ),
  requestCameraPermissionsAsync: jest.fn(() =>
    Promise.resolve({ status: 'granted' })
  ),
  MediaTypeOptions: { Images: 'Images', Videos: 'Videos', All: 'All' },
}));

// expo-document-picker mock
jest.mock('expo-document-picker', () => ({
  getDocumentAsync: jest.fn(),
}));

// i18next mock
jest.mock('i18next', () => ({
  use: jest.fn().mockReturnThis(),
  init: jest.fn().mockReturnThis(),
  t: jest.fn((key) => key),
  language: 'en',
  changeLanguage: jest.fn(() => Promise.resolve()),
}));

// react-i18next mock
jest.mock('react-i18next', () => ({
  useTranslation: jest.fn(() => ({
    t: jest.fn((key) => key),
    i18n: {
      language: 'en',
      changeLanguage: jest.fn(() => Promise.resolve()),
    },
  })),
  initReactI18next: { type: '3rdParty', init: jest.fn() },
}));
