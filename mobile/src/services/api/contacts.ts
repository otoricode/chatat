// Contacts API service (placeholder)
import apiClient from './client';

type Contact = {
  id: string;
  name: string;
  phone: string;
  avatar: string;
  status: string;
  isOnline: boolean;
  lastSeen: string | null;
};

type SyncContactsRequest = {
  phoneHashes: string[];
};

type SyncContactsResponse = {
  matched: number;
};

export const contactsApi = {
  sync: (data: SyncContactsRequest) =>
    apiClient.post<SyncContactsResponse>('/contacts/sync', data),

  list: () => apiClient.get<Contact[]>('/contacts'),

  search: (phone: string) =>
    apiClient.get<Contact[]>(`/contacts/search?phone=${encodeURIComponent(phone)}`),

  getProfile: (userId: string) =>
    apiClient.get<Contact>(`/contacts/${userId}`),
};
