// Block component props shared across all block types
import type { Block, BlockType } from '@/types/chat';

export interface BlockProps {
  block: Block;
  isActive: boolean;
  readOnly: boolean;
  onChange: (changes: Partial<Block>) => void;
  onFocus: () => void;
  onSubmit: () => void;
  onBackspace: () => void;
  onSlashTrigger: () => void;
}

export interface SlashOption {
  type: BlockType;
  icon: string;
  label: string;
  description: string;
}

export const BLOCK_OPTIONS: SlashOption[] = [
  { type: 'paragraph', icon: '📝', label: 'Teks', description: 'Teks biasa' },
  { type: 'heading1', icon: '𝗛₁', label: 'Judul 1', description: 'Judul besar' },
  { type: 'heading2', icon: '𝗛₂', label: 'Judul 2', description: 'Judul sedang' },
  { type: 'heading3', icon: '𝗛₃', label: 'Judul 3', description: 'Judul kecil' },
  { type: 'bullet-list', icon: '•', label: 'Daftar Bullet', description: 'Daftar tak berurutan' },
  { type: 'numbered-list', icon: '1.', label: 'Daftar Nomor', description: 'Daftar berurutan' },
  { type: 'checklist', icon: '☑', label: 'Checklist', description: 'Daftar centang' },
  { type: 'table', icon: '▦', label: 'Tabel', description: 'Tabel data' },
  { type: 'callout', icon: '💡', label: 'Callout', description: 'Blok perhatian' },
  { type: 'code', icon: '⌨', label: 'Kode', description: 'Blok kode' },
  { type: 'toggle', icon: '▶', label: 'Toggle', description: 'Blok toggle' },
  { type: 'divider', icon: '—', label: 'Pembatas', description: 'Garis horizontal' },
  { type: 'quote', icon: '❝', label: 'Kutipan', description: 'Blok kutipan' },
];
