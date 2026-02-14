import { useSettingsStore } from '../settingsStore';

const defaultNotifications = {
  showPreview: true,
  soundEnabled: true,
  vibrationEnabled: true,
  groupAlerts: true,
};

const defaultAutoDownload = {
  wifiAll: true,
  cellularUnder5MB: true,
  cellularAsk: true,
};

beforeEach(() => {
  useSettingsStore.getState().reset();
});

describe('settingsStore', () => {
  it('starts with default notification settings', () => {
    const state = useSettingsStore.getState();
    expect(state.notifications).toEqual(defaultNotifications);
  });

  it('starts with default auto-download settings', () => {
    const state = useSettingsStore.getState();
    expect(state.autoDownload).toEqual(defaultAutoDownload);
  });

  it('updateNotifications merges partial updates', () => {
    useSettingsStore.getState().updateNotifications({ soundEnabled: false });

    const state = useSettingsStore.getState();
    expect(state.notifications.soundEnabled).toBe(false);
    expect(state.notifications.showPreview).toBe(true); // unchanged
    expect(state.notifications.vibrationEnabled).toBe(true); // unchanged
  });

  it('updateAutoDownload merges partial updates', () => {
    useSettingsStore.getState().updateAutoDownload({ wifiAll: false });

    const state = useSettingsStore.getState();
    expect(state.autoDownload.wifiAll).toBe(false);
    expect(state.autoDownload.cellularUnder5MB).toBe(true); // unchanged
  });

  it('reset restores defaults', () => {
    useSettingsStore.getState().updateNotifications({ soundEnabled: false, showPreview: false });
    useSettingsStore.getState().updateAutoDownload({ wifiAll: false });
    useSettingsStore.getState().reset();

    const state = useSettingsStore.getState();
    expect(state.notifications).toEqual(defaultNotifications);
    expect(state.autoDownload).toEqual(defaultAutoDownload);
  });
});
