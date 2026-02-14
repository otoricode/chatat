// @ts-nocheck
import { useAuthStore } from '../authStore';

// Reset store before each test
beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: false,
    accessToken: null,
    refreshToken: null,
    user: null,
    isNewUser: false,
  });
});

describe('authStore', () => {
  it('starts unauthenticated', () => {
    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.accessToken).toBeNull();
    expect(state.refreshToken).toBeNull();
    expect(state.user).toBeNull();
  });

  it('login sets auth state', () => {
    const tokens = { accessToken: 'access123', refreshToken: 'refresh123' };
    const user = { id: '1', name: 'Test', phone: '+62', avatar: '', status: '' };

    useAuthStore.getState().login(tokens, user, false);

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.accessToken).toBe('access123');
    expect(state.refreshToken).toBe('refresh123');
    expect(state.user?.name).toBe('Test');
    expect(state.isNewUser).toBe(false);
  });

  it('login with isNewUser true', () => {
    const tokens = { accessToken: 'a', refreshToken: 'r' };
    const user = { id: '1', name: 'New', phone: '+62', avatar: '', status: '' };

    useAuthStore.getState().login(tokens, user, true);

    expect(useAuthStore.getState().isNewUser).toBe(true);
  });

  it('logout clears all state', () => {
    const tokens = { accessToken: 'a', refreshToken: 'r' };
    const user = { id: '1', name: 'Test', phone: '+62', avatar: '', status: '' };

    useAuthStore.getState().login(tokens, user, false);
    useAuthStore.getState().logout();

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.accessToken).toBeNull();
    expect(state.refreshToken).toBeNull();
    expect(state.user).toBeNull();
  });

  it('setUser updates user', () => {
    const user = { id: '2', name: 'New User', phone: '+1', avatar: 'av', status: 'hello' };

    useAuthStore.getState().setUser(user);

    expect(useAuthStore.getState().user?.name).toBe('New User');
  });

  it('updateProfile merges partial updates', () => {
    const user = { id: '1', name: 'Old', phone: '+62', avatar: '', status: 'busy' };

    useAuthStore.getState().setUser(user);
    useAuthStore.getState().updateProfile({ name: 'Updated' });

    const updated = useAuthStore.getState().user;
    expect(updated?.name).toBe('Updated');
    expect(updated?.phone).toBe('+62');
    expect(updated?.status).toBe('busy');
  });

  it('updateProfile does nothing when user is null', () => {
    useAuthStore.getState().updateProfile({ name: 'X' });

    expect(useAuthStore.getState().user).toBeNull();
  });

  it('setTokens updates token pair', () => {
    useAuthStore.getState().setTokens({
      accessToken: 'new-access',
      refreshToken: 'new-refresh',
    });

    const state = useAuthStore.getState();
    expect(state.accessToken).toBe('new-access');
    expect(state.refreshToken).toBe('new-refresh');
  });
});
