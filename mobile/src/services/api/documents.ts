// Documents API service
import apiClient from './client';
import type {
  Block,
  BlockType,
  CollaboratorRole,
  DocumentFull,
  DocumentListItem,
  DocumentTemplate,
  DocumentHistory,
  DocumentSigner,
} from '@/types/chat';

export interface CreateDocumentInput {
  title: string;
  icon?: string;
  chatId?: string;
  topicId?: string;
  isStandalone?: boolean;
  templateId?: string;
}

export interface UpdateDocumentInput {
  title?: string;
  icon?: string;
  cover?: string;
}

export const documentsApi = {
  // Document CRUD
  create: (data: CreateDocumentInput) =>
    apiClient.post<{ data: DocumentFull }>('/documents', data),

  getById: (id: string) =>
    apiClient.get<{ data: DocumentFull }>(`/documents/${id}`),

  list: (cursor?: string, limit?: number) =>
    apiClient.get<{ data: DocumentListItem[]; meta: { cursor: string; hasMore: boolean } }>(
      '/documents',
      { params: { cursor, limit } },
    ),

  listByChat: (chatId: string) =>
    apiClient.get<{ data: DocumentListItem[] }>(`/chats/${chatId}/documents`),

  listByTopic: (topicId: string) =>
    apiClient.get<{ data: DocumentListItem[] }>(`/topics/${topicId}/documents`),

  update: (id: string, data: UpdateDocumentInput) =>
    apiClient.put<{ data: DocumentFull }>(`/documents/${id}`, data),

  delete: (id: string) =>
    apiClient.delete(`/documents/${id}`),

  duplicate: (id: string) =>
    apiClient.post<{ data: DocumentFull }>(`/documents/${id}/duplicate`),

  // Blocks
  addBlock: (docId: string, data: { type: BlockType; content?: string; sortOrder: number }) =>
    apiClient.post<{ data: Block }>(`/documents/${docId}/blocks`, data),

  updateBlock: (docId: string, blockId: string, data: Partial<Block>) =>
    apiClient.put<{ data: Block }>(`/documents/${docId}/blocks/${blockId}`, data),

  deleteBlock: (docId: string, blockId: string) =>
    apiClient.delete(`/documents/${docId}/blocks/${blockId}`),

  reorderBlocks: (docId: string, blockIds: string[]) =>
    apiClient.put(`/documents/${docId}/blocks/reorder`, { blockIds }),

  batchBlocks: (docId: string, operations: { action: string; block?: Partial<Block> }[]) =>
    apiClient.post(`/documents/${docId}/blocks/batch`, { operations }),

  // Collaborators
  addCollaborator: (docId: string, userId: string, role: CollaboratorRole) =>
    apiClient.post(`/documents/${docId}/collaborators`, { userId, role }),

  removeCollaborator: (docId: string, userId: string) =>
    apiClient.delete(`/documents/${docId}/collaborators/${userId}`),

  updateCollaboratorRole: (docId: string, userId: string, role: CollaboratorRole) =>
    apiClient.put(`/documents/${docId}/collaborators/${userId}`, { role }),

  // Tags
  addTag: (docId: string, tag: string) =>
    apiClient.post(`/documents/${docId}/tags`, { tag }),

  removeTag: (docId: string, tag: string) =>
    apiClient.delete(`/documents/${docId}/tags/${tag}`),

  // History
  getHistory: (docId: string) =>
    apiClient.get<{ data: DocumentHistory[] }>(`/documents/${docId}/history`),

  // Templates
  getTemplates: () =>
    apiClient.get<{ data: DocumentTemplate[] }>('/templates'),

  // Lock & Sign
  lock: (id: string, mode: 'manual' | 'signatures') =>
    apiClient.post(`/documents/${id}/lock`, { mode }),

  unlock: (id: string) =>
    apiClient.post(`/documents/${id}/unlock`),

  sign: (id: string, name?: string) =>
    apiClient.post(`/documents/${id}/sign`, { name }),

  // Signers
  listSigners: (docId: string) =>
    apiClient.get<{ data: DocumentSigner[] }>(`/documents/${docId}/signers`),

  addSigner: (docId: string, userId: string) =>
    apiClient.post(`/documents/${docId}/signers`, { userId }),

  removeSigner: (docId: string, userId: string) =>
    apiClient.delete(`/documents/${docId}/signers/${userId}`),
};
