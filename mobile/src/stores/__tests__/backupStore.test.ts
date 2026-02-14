jest.mock('@/services/api/backup', () => ({
  backupApi: {
    getLatest: jest.fn(),
    getHistory: jest.fn(),
  },
}));

jest.mock('@/services/backup/CloudBackupService', () => ({
  performBackup: jest.fn(),
  performRestore: jest.fn(),
  listCloudBackups: jest.fn(),
  getPlatform: jest.fn(() => 'ios'),
}));

import { useBackupStore } from '../backupStore';
import { backupApi } from '@/services/api/backup';
import { performBackup, performRestore, listCloudBackups } from '@/services/backup/CloudBackupService';

const mockBackupApi = backupApi as jest.Mocked<typeof backupApi>;
const mockPerformBackup = performBackup as jest.MockedFunction<typeof performBackup>;
const mockPerformRestore = performRestore as jest.MockedFunction<typeof performRestore>;
const mockListCloud = listCloudBackups as jest.MockedFunction<typeof listCloudBackups>;

const makeBackupRecord = (id: string) => ({
  id,
  userId: 'u1',
  size: 1024,
  createdAt: '2024-01-01T00:00:00Z',
});

beforeEach(() => {
  useBackupStore.setState({
    latestBackup: null,
    history: [],
    cloudBackups: [],
    isLoading: false,
    isBackingUp: false,
    isRestoring: false,
    progress: null,
    error: null,
    platform: 'ios',
  });
  jest.clearAllMocks();
});

describe('backupStore', () => {
  it('starts with defaults', () => {
    const s = useBackupStore.getState();
    expect(s.latestBackup).toBeNull();
    expect(s.history).toEqual([]);
    expect(s.isBackingUp).toBe(false);
    expect(s.isRestoring).toBe(false);
    expect(s.platform).toBe('ios');
  });

  it('fetchLatest success', async () => {
    const record = makeBackupRecord('b1');
    mockBackupApi.getLatest.mockResolvedValue({
      data: { success: true, data: record },
    } as any);

    await useBackupStore.getState().fetchLatest();

    expect(useBackupStore.getState().latestBackup).toEqual(record);
  });

  it('fetchLatest handles null data', async () => {
    mockBackupApi.getLatest.mockResolvedValue({
      data: { success: true, data: null },
    } as any);

    await useBackupStore.getState().fetchLatest();

    expect(useBackupStore.getState().latestBackup).toBeNull();
  });

  it('fetchLatest error with Error', async () => {
    mockBackupApi.getLatest.mockRejectedValue(new Error('Failed'));

    await useBackupStore.getState().fetchLatest();

    expect(useBackupStore.getState().error).toBe('Failed');
  });

  it('fetchLatest error with non-Error', async () => {
    mockBackupApi.getLatest.mockRejectedValue('err');

    await useBackupStore.getState().fetchLatest();

    expect(useBackupStore.getState().error).toBe('Failed to fetch latest backup');
  });

  it('fetchHistory success', async () => {
    const records = [makeBackupRecord('b1'), makeBackupRecord('b2')];
    mockBackupApi.getHistory.mockResolvedValue({
      data: { success: true, data: records },
    } as any);

    await useBackupStore.getState().fetchHistory();

    expect(useBackupStore.getState().history).toHaveLength(2);
    expect(useBackupStore.getState().isLoading).toBe(false);
  });

  it('fetchHistory handles null data', async () => {
    mockBackupApi.getHistory.mockResolvedValue({
      data: { success: true, data: null },
    } as any);

    await useBackupStore.getState().fetchHistory();

    expect(useBackupStore.getState().history).toEqual([]);
  });

  it('fetchHistory error with Error', async () => {
    mockBackupApi.getHistory.mockRejectedValue(new Error('History fail'));

    await useBackupStore.getState().fetchHistory();

    expect(useBackupStore.getState().error).toBe('History fail');
    expect(useBackupStore.getState().isLoading).toBe(false);
  });

  it('fetchHistory error with non-Error', async () => {
    mockBackupApi.getHistory.mockRejectedValue(42);

    await useBackupStore.getState().fetchHistory();

    expect(useBackupStore.getState().error).toBe('Failed to fetch backup history');
  });

  it('fetchCloudBackups success', async () => {
    const files = [{ name: 'backup1.zip', size: 1024, createdAt: '2024-01-01' }];
    mockListCloud.mockResolvedValue(files as any);

    await useBackupStore.getState().fetchCloudBackups();

    expect(useBackupStore.getState().cloudBackups).toEqual(files);
  });

  it('fetchCloudBackups error is silent', async () => {
    mockListCloud.mockRejectedValue(new Error('Not signed in'));

    await useBackupStore.getState().fetchCloudBackups();

    // No error set
    expect(useBackupStore.getState().error).toBeNull();
  });

  it('startBackup success', async () => {
    mockPerformBackup.mockResolvedValue(undefined as any);
    mockBackupApi.getLatest.mockResolvedValue({
      data: { success: true, data: makeBackupRecord('b1') },
    } as any);

    await useBackupStore.getState().startBackup();

    expect(useBackupStore.getState().isBackingUp).toBe(false);
    expect(useBackupStore.getState().latestBackup).toBeTruthy();
  });

  it('startBackup error with Error', async () => {
    mockPerformBackup.mockRejectedValue(new Error('Backup failed'));

    await useBackupStore.getState().startBackup();

    expect(useBackupStore.getState().error).toBe('Backup failed');
    expect(useBackupStore.getState().isBackingUp).toBe(false);
  });

  it('startBackup error with non-Error', async () => {
    mockPerformBackup.mockRejectedValue(42);

    await useBackupStore.getState().startBackup();

    expect(useBackupStore.getState().error).toBe('Backup failed');
  });

  it('startRestore success', async () => {
    mockPerformRestore.mockResolvedValue(undefined as any);

    await useBackupStore.getState().startRestore();

    expect(useBackupStore.getState().isRestoring).toBe(false);
    expect(useBackupStore.getState().progress).toBeNull();
  });

  it('startRestore error with Error', async () => {
    mockPerformRestore.mockRejectedValue(new Error('Restore failed'));

    await useBackupStore.getState().startRestore();

    expect(useBackupStore.getState().error).toBe('Restore failed');
    expect(useBackupStore.getState().isRestoring).toBe(false);
  });

  it('startRestore error with non-Error', async () => {
    mockPerformRestore.mockRejectedValue('err');

    await useBackupStore.getState().startRestore();

    expect(useBackupStore.getState().error).toBe('Restore failed');
  });

  it('clearError resets error', () => {
    useBackupStore.setState({ error: 'some error' });

    useBackupStore.getState().clearError();

    expect(useBackupStore.getState().error).toBeNull();
  });
});
