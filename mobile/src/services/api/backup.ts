// Backup API service
import apiClient from './client';

export type BackupPlatform = 'google_drive' | 'icloud';
export type BackupStatus = 'in_progress' | 'completed' | 'failed';

export type BackupRecord = {
  id: string;
  userId: string;
  sizeBytes: number;
  platform: BackupPlatform;
  status: BackupStatus;
  metadata?: Record<string, unknown>;
  createdAt: string;
};

export type BackupBundle = {
  version: number;
  userId: string;
  createdAt: string;
  data: BackupData;
};

export type BackupData = {
  profile: UserExport | null;
  chats: ChatExport[];
  messages: MessageExport[];
  contacts: ContactExport[];
  documents: DocumentExport[];
};

export type UserExport = {
  id: string;
  name: string;
  phone: string;
  avatar: string;
  about: string;
};

export type ChatExport = {
  id: string;
  type: string;
  name: string;
  createdAt: string;
};

export type MessageExport = {
  id: string;
  chatId: string;
  senderName: string;
  content: string;
  contentType: string;
  createdAt: string;
};

export type ContactExport = {
  userId: string;
  name: string;
  phone: string;
};

export type DocumentExport = {
  id: string;
  title: string;
  docType: string;
  createdAt: string;
};

export type LogBackupInput = {
  sizeBytes: number;
  platform: BackupPlatform;
  status: BackupStatus;
};

export const backupApi = {
  export: () => apiClient.get<BackupBundle>('/backup/export'),

  import: (bundle: BackupBundle) =>
    apiClient.post<{ message: string }>('/backup/import', bundle),

  log: (input: LogBackupInput) =>
    apiClient.post<BackupRecord>('/backup/log', input),

  getHistory: () => apiClient.get<BackupRecord[]>('/backup/history'),

  getLatest: () => apiClient.get<BackupRecord | null>('/backup/latest'),
};
