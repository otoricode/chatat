// Entities API service
import apiClient from './client';
import type { Entity, EntityListItem, EntityListResponse, Document } from '@/types/chat';

export interface CreateEntityInput {
  name: string;
  type: string;
  fields?: Record<string, string>;
}

export interface UpdateEntityInput {
  name?: string;
  type?: string;
  fields?: Record<string, string>;
}

export const entitiesApi = {
  // CRUD
  create: (data: CreateEntityInput) =>
    apiClient.post<{ data: Entity }>('/entities', data),

  list: (params?: { type?: string; limit?: number; offset?: number }) =>
    apiClient.get<{ data: EntityListResponse }>('/entities', { params }),

  getById: (id: string) =>
    apiClient.get<{ data: Entity }>(`/entities/${id}`),

  update: (id: string, data: UpdateEntityInput) =>
    apiClient.put<{ data: Entity }>(`/entities/${id}`, data),

  delete: (id: string) =>
    apiClient.delete(`/entities/${id}`),

  search: (q: string) =>
    apiClient.get<{ data: Entity[] }>('/entities/search', { params: { q } }),

  listTypes: () =>
    apiClient.get<{ data: string[] }>('/entities/types'),

  // Document linking
  listDocuments: (entityId: string) =>
    apiClient.get<{ data: Document[] }>(`/entities/${entityId}/documents`),

  linkToDocument: (documentId: string, entityId: string) =>
    apiClient.post(`/documents/${documentId}/entities`, { entityId }),

  unlinkFromDocument: (documentId: string, entityId: string) =>
    apiClient.delete(`/documents/${documentId}/entities/${entityId}`),

  getDocumentEntities: (documentId: string) =>
    apiClient.get<{ data: Entity[] }>(`/documents/${documentId}/entities`),

  // Contact as entity
  createFromContact: (contactUserId: string) =>
    apiClient.post<{ data: Entity }>('/entities/from-contact', { contactUserId }),
};
