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
