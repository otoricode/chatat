// Navigation type definitions for Chatat
// Based on spesifikasi-chatat.md section 7 and plan/phase-06-mobile-shell.md

import type { NavigatorScreenParams } from '@react-navigation/native';

// Root level: Auth or Main
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
};

// Auth flow screens
export type AuthStackParamList = {
  PhoneInput: undefined;
  OTPVerify: { phone: string; method: 'sms' | 'reverse' };
  ReverseOTPWait: { sessionId: string; waNumber: string; code: string };
  ProfileSetup: undefined;
};

// Main bottom tabs: Chat and Document
export type MainTabParamList = {
  ChatTab: NavigatorScreenParams<ChatStackParamList>;
  DocumentTab: NavigatorScreenParams<DocumentStackParamList>;
};

// Chat stack (nested in ChatTab)
export type ChatStackParamList = {
  ChatList: undefined;
  Chat: { chatId: string; chatType: 'personal' | 'group' };
  ChatInfo: { chatId: string; chatType: 'personal' | 'group' };
  ContactList: undefined;
  CreateGroup: undefined;
  TopicList: { chatId: string };
  CreateTopic: { chatId: string; chatType: 'personal' | 'group' };
  Topic: { topicId: string };
  TopicInfo: { topicId: string };
  ImageViewer: { url: string; filename?: string };
  DocumentEditor: {
    documentId?: string;
    contextType?: string;
    contextId?: string;
  };
};

// Document stack (nested in DocumentTab)
export type DocumentStackParamList = {
  DocumentList: undefined;
  DocumentEditor: {
    documentId?: string;
    contextType?: string;
    contextId?: string;
  };
  DocumentViewer: { documentId: string };
  EntityList: undefined;
  EntityDetail: { entityId: string };
};
