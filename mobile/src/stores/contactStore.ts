// Contact store — manages contacts state
import { create } from 'zustand';
import { contactsApi } from '@/services/api/contacts';
import type { ContactInfo } from '@/types/chat';

type ContactState = {
  contacts: ContactInfo[];
  isLoading: boolean;
  lastSynced: string | null;
  error: string | null;

  fetchContacts: () => Promise<void>;
  searchContacts: (query: string) => ContactInfo[];
  updateOnlineStatus: (userId: string, isOnline: boolean, lastSeen: string) => void;
  clearError: () => void;
};

export const useContactStore = create<ContactState>()((set, get) => ({
  contacts: [],
  isLoading: false,
  lastSynced: null,
  error: null,

  fetchContacts: async () => {
    set({ isLoading: true, error: null });
    try {
      const res = await contactsApi.list();
      const data = res.data as unknown as { success: boolean; data: ContactInfo[] };
      set({
        contacts: data.data ?? [],
        isLoading: false,
        lastSynced: new Date().toISOString(),
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to load contacts';
      set({ error: msg, isLoading: false });
    }
  },

  searchContacts: (query) => {
    const { contacts } = get();
    if (!query.trim()) return contacts;
    const lower = query.toLowerCase();
    return contacts.filter(
      (c) =>
        c.name.toLowerCase().includes(lower) || c.phone.includes(query),
    );
  },

  clearError: () => set({ error: null }),

  updateOnlineStatus: (userId, isOnline, lastSeen) => {
    set((state) => ({
      contacts: state.contacts.map((c) =>
        c.userId === userId ? { ...c, isOnline, lastSeen } : c,
      ),
    }));
  },
}));
