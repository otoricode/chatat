// Search API service
import apiClient from './client';

export interface MessageSearchResult {
  id: string;
  chatId: string;
  senderId: string;
  content: string;
  type: string;
  createdAt: string;
  chatName: string;
  senderName: string;
  highlight: string;
}

export interface DocumentSearchResult {
  id: string;
  title: string;
  icon: string;
  ownerId: string;
  locked: boolean;
  updatedAt: string;
  highlight: string;
}

export interface ContactSearchResult {
  id: string;
  phone: string;
  name: string;
  avatar: string;
  status: string;
  lastSeen: string;
}

export interface EntitySearchResult {
  id: string;
  name: string;
  type: string;
  fields: Record<string, string>;
  ownerId: string;
  contactUserId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SearchAllResponse {
  data: {
    messages: MessageSearchResult[];
    documents: DocumentSearchResult[];
    contacts: ContactSearchResult[];
    entities: EntitySearchResult[];
  };
}

export const searchApi = {
  searchAll: (q: string, limit = 3) =>
    apiClient.get<SearchAllResponse>('/search', { params: { q, limit } }),

  searchMessages: (q: string, offset = 0, limit = 20) =>
    apiClient.get<{ data: MessageSearchResult[] }>('/search/messages', {
      params: { q, offset, limit },
    }),

  searchDocuments: (q: string, offset = 0, limit = 20) =>
    apiClient.get<{ data: DocumentSearchResult[] }>('/search/documents', {
      params: { q, offset, limit },
    }),

  searchContacts: (q: string) =>
    apiClient.get<{ data: ContactSearchResult[] }>('/search/contacts', {
      params: { q },
    }),

  searchEntities: (q: string) =>
    apiClient.get<{ data: EntitySearchResult[] }>('/search/entities', {
      params: { q },
    }),

  searchInChat: (chatId: string, q: string, offset = 0, limit = 20) =>
    apiClient.get<{ data: MessageSearchResult[] }>(`/chats/${chatId}/search`, {
      params: { q, offset, limit },
    }),
};
