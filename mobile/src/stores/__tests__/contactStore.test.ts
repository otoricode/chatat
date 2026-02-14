// @ts-nocheck
jest.mock('@/services/api/contacts', () => ({
  contactsApi: {
    list: jest.fn(),
  },
}));

import { useContactStore } from '../contactStore';
import { contactsApi } from '@/services/api/contacts';

const mockContactsApi = contactsApi as jest.Mocked<typeof contactsApi>;

const makeContact = (userId: string, name: string, phone: string) => ({
  userId,
  name,
  phone,
  avatar: '',
  isOnline: false,
  lastSeen: '',
});

beforeEach(() => {
  useContactStore.setState({
    contacts: [],
    isLoading: false,
    lastSynced: null,
    error: null,
  });
  jest.clearAllMocks();
});

describe('contactStore', () => {
  it('starts empty', () => {
    const s = useContactStore.getState();
    expect(s.contacts).toEqual([]);
    expect(s.isLoading).toBe(false);
    expect(s.lastSynced).toBeNull();
  });

  it('fetchContacts success', async () => {
    const data = [makeContact('u1', 'Alice', '+62111'), makeContact('u2', 'Bob', '+62222')];
    mockContactsApi.list.mockResolvedValue({
      data: { success: true, data },
    } as any);

    await useContactStore.getState().fetchContacts();

    const s = useContactStore.getState();
    expect(s.contacts).toHaveLength(2);
    expect(s.isLoading).toBe(false);
    expect(s.lastSynced).toBeTruthy();
  });

  it('fetchContacts handles null data', async () => {
    mockContactsApi.list.mockResolvedValue({
      data: { success: true, data: null },
    } as any);

    await useContactStore.getState().fetchContacts();

    expect(useContactStore.getState().contacts).toEqual([]);
  });

  it('fetchContacts error with Error instance', async () => {
    mockContactsApi.list.mockRejectedValue(new Error('Network error'));

    await useContactStore.getState().fetchContacts();

    const s = useContactStore.getState();
    expect(s.error).toBe('Network error');
    expect(s.isLoading).toBe(false);
  });

  it('fetchContacts error with non-Error', async () => {
    mockContactsApi.list.mockRejectedValue('fail');

    await useContactStore.getState().fetchContacts();

    expect(useContactStore.getState().error).toBe('Failed to load contacts');
  });

  it('searchContacts returns all for empty query', () => {
    const data = [makeContact('u1', 'Alice', '+62111'), makeContact('u2', 'Bob', '+62222')];
    useContactStore.setState({ contacts: data as any });

    const result = useContactStore.getState().searchContacts('');
    expect(result).toHaveLength(2);
  });

  it('searchContacts filters by name', () => {
    const data = [makeContact('u1', 'Alice', '+62111'), makeContact('u2', 'Bob', '+62222')];
    useContactStore.setState({ contacts: data as any });

    const result = useContactStore.getState().searchContacts('ali');
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Alice');
  });

  it('searchContacts filters by phone', () => {
    const data = [makeContact('u1', 'Alice', '+62111'), makeContact('u2', 'Bob', '+62222')];
    useContactStore.setState({ contacts: data as any });

    const result = useContactStore.getState().searchContacts('+62222');
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Bob');
  });

  it('searchContacts returns empty for no match', () => {
    const data = [makeContact('u1', 'Alice', '+62111')];
    useContactStore.setState({ contacts: data as any });

    const result = useContactStore.getState().searchContacts('XYZ');
    expect(result).toHaveLength(0);
  });

  it('clearError resets error', () => {
    useContactStore.setState({ error: 'some error' });

    useContactStore.getState().clearError();

    expect(useContactStore.getState().error).toBeNull();
  });

  it('updateOnlineStatus updates matching user', () => {
    const data = [makeContact('u1', 'Alice', '+62111')];
    useContactStore.setState({ contacts: data as any });

    useContactStore.getState().updateOnlineStatus('u1', true, '2024-01-01T12:00:00Z');

    const c = useContactStore.getState().contacts[0] as any;
    expect(c.isOnline).toBe(true);
    expect(c.lastSeen).toBe('2024-01-01T12:00:00Z');
  });

  it('updateOnlineStatus ignores non-matching user', () => {
    const data = [makeContact('u1', 'Alice', '+62111')];
    useContactStore.setState({ contacts: data as any });

    useContactStore.getState().updateOnlineStatus('u999', true, '2024-01-01');

    expect((useContactStore.getState().contacts[0] as any).isOnline).toBe(false);
  });
});
