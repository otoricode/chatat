// Chat and message types matching backend models

export type ChatType = 'personal' | 'group';
export type MessageType = 'text' | 'image' | 'file' | 'document_card' | 'system';
export type DeliveryStatus = 'sent' | 'delivered' | 'read';
export type MemberRole = 'admin' | 'member';

export interface User {
  id: string;
  name: string;
  phone: string;
  avatar: string;
  status: string;
  lastSeen: string;
}

export interface Chat {
  id: string;
  type: ChatType;
  name: string;
  icon: string;
  description: string;
  createdBy: string;
  pinnedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Message {
  id: string;
  chatId: string;
  senderId: string;
  content: string;
  replyToId: string | null;
  type: MessageType;
  metadata: Record<string, unknown> | null;
  isDeleted: boolean;
  deletedForAll: boolean;
  createdAt: string;
}

export interface ChatListItem {
  chat: Chat;
  lastMessage: Message | null;
  unreadCount: number;
  otherUser: User | null;
  isOnline: boolean;
}

export interface ChatDetail {
  chat: Chat;
  members: User[];
}

export interface MemberInfo {
  user: User;
  role: MemberRole;
  isOnline: boolean;
  joinedAt: string;
}

export interface GroupInfo {
  chat: Chat;
  members: MemberInfo[];
}

export interface MessagePage {
  messages: Message[];
  cursor: string;
  hasMore: boolean;
}

export interface ContactInfo {
  userId: string;
  phone: string;
  name: string;
  avatar: string;
  status: string;
  isOnline: boolean;
  lastSeen: string;
}

// API response wrappers
export interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export interface PaginatedResponse<T> {
  success: boolean;
  data: T;
  meta: {
    cursor: string;
    hasMore: boolean;
    total?: number;
  };
}

// Topic types
export interface Topic {
  id: string;
  name: string;
  icon: string;
  description: string;
  parentType: ChatType;
  parentId: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface TopicMessage {
  id: string;
  topicId: string;
  senderId: string;
  content: string;
  replyToId: string | null;
  type: MessageType;
  isDeleted: boolean;
  deletedForAll: boolean;
  createdAt: string;
}

export interface TopicListItem {
  topic: Topic;
  lastMessage: TopicMessage | null;
  memberCount: number;
}

export interface TopicDetail {
  topic: Topic;
  members: MemberInfo[];
  parent: Chat | null;
}

// Media types
export type MediaType = 'image' | 'file';

export interface MediaResponse {
  id: string;
  type: MediaType;
  filename: string;
  contentType: string;
  size: number;
  width?: number;
  height?: number;
  url: string;
  thumbnailURL: string;
  createdAt: string;
}

// Document types
export type BlockType =
  | 'paragraph'
  | 'heading1'
  | 'heading2'
  | 'heading3'
  | 'bullet-list'
  | 'numbered-list'
  | 'checklist'
  | 'table'
  | 'callout'
  | 'code'
  | 'toggle'
  | 'divider'
  | 'quote';

export type CollaboratorRole = 'editor' | 'viewer';

export interface Block {
  id: string;
  documentId: string;
  type: BlockType;
  content: string;
  checked?: boolean;
  rows?: unknown[];
  columns?: unknown[];
  language?: string;
  emoji?: string;
  color?: string;
  sortOrder: number;
  parentBlockId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Document {
  id: string;
  title: string;
  icon: string;
  cover?: string;
  ownerId: string;
  chatId?: string;
  topicId?: string;
  isStandalone: boolean;
  requireSigs: boolean;
  locked: boolean;
  lockedAt?: string;
  lockedBy?: string;
  createdAt: string;
  updatedAt: string;
}

export interface DocumentCollaboratorInfo {
  userId: string;
  name: string;
  avatar: string;
  role: CollaboratorRole;
  addedAt: string;
}

export interface DocumentSigner {
  documentId: string;
  userId: string;
  signedAt?: string;
  signerName: string;
}

export interface DocumentFull {
  document: Document;
  blocks: Block[];
  collaborators: DocumentCollaboratorInfo[];
  signers: DocumentSigner[];
  tags: string[];
  history: DocumentHistory[];
}

export interface DocumentListItem {
  id: string;
  title: string;
  icon: string;
  locked: boolean;
  requireSigs: boolean;
  ownerId: string;
  contextType: string;
  updatedAt: string;
}

export interface DocumentHistory {
  id: string;
  documentId: string;
  userId: string;
  action: string;
  details: string;
  createdAt: string;
}

export interface DocumentTemplate {
  id: string;
  name: string;
  icon: string;
  blocks: TemplateBlock[];
}

export interface TemplateBlock {
  type: string;
  content: string;
  rows?: unknown;
  columns?: unknown;
  emoji?: string;
  color?: string;
}
