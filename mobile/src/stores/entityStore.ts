// Entity store — manages entity state
import { create } from 'zustand';
import i18n from 'i18next';
import { entitiesApi } from '@/services/api/entities';
import type { Entity, EntityListItem } from '@/types/chat';

type EntityState = {
  entities: EntityListItem[];
  types: string[];
  isLoading: boolean;
  total: number;
  error: string | null;

  fetchEntities: (type?: string, limit?: number, offset?: number) => Promise<void>;
  fetchTypes: () => Promise<void>;
  searchEntities: (query: string) => Promise<Entity[]>;
  createEntity: (input: { name: string; type: string; fields?: Record<string, string> }) => Promise<Entity>;
  updateEntity: (id: string, input: { name?: string; type?: string; fields?: Record<string, string> }) => Promise<Entity>;
  deleteEntity: (id: string) => Promise<void>;
  createFromContact: (contactUserId: string) => Promise<Entity>;
  clearError: () => void;
};

export const useEntityStore = create<EntityState>()((set) => ({
  entities: [],
  types: [],
  isLoading: false,
  total: 0,
  error: null,

  fetchEntities: async (type, limit = 20, offset = 0) => {
    set({ isLoading: true, error: null });
    try {
      const res = await entitiesApi.list({ type, limit, offset });
      const payload = res.data as unknown as { data: { data: EntityListItem[]; total: number; limit: number; offset: number } };
      set({
        entities: payload.data.data ?? [],
        total: payload.data.total ?? 0,
        isLoading: false,
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : i18n.t('entity.loadFailed');
      set({ error: msg, isLoading: false });
    }
  },

  fetchTypes: async () => {
    try {
      const res = await entitiesApi.listTypes();
      const payload = res.data as unknown as { data: string[] };
      set({ types: payload.data ?? [] });
    } catch {
      // silently ignore
    }
  },

  searchEntities: async (query: string) => {
    try {
      const res = await entitiesApi.search(query);
      const payload = res.data as unknown as { data: Entity[] };
      return payload.data ?? [];
    } catch {
      return [];
    }
  },

  createEntity: async (input) => {
    const res = await entitiesApi.create(input);
    const payload = res.data as unknown as { data: Entity };
    return payload.data;
  },

  updateEntity: async (id, input) => {
    const res = await entitiesApi.update(id, input);
    const payload = res.data as unknown as { data: Entity };
    return payload.data;
  },

  deleteEntity: async (id) => {
    await entitiesApi.delete(id);
    set((state) => ({
      entities: state.entities.filter((e) => e.id !== id),
    }));
  },

  createFromContact: async (contactUserId) => {
    const res = await entitiesApi.createFromContact(contactUserId);
    const payload = res.data as unknown as { data: Entity };
    return payload.data;
  },

  clearError: () => set({ error: null }),
}));
