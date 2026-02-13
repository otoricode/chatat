// Editor store — manages block editor state
import { create } from 'zustand';
import i18n from 'i18next';
import { documentsApi } from '@/services/api/documents';
import type { Block, BlockType, DocumentFull } from '@/types/chat';

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error';

type EditorState = {
  documentId: string | null;
  title: string;
  icon: string;
  blocks: Block[];
  activeBlockId: string | null;
  isEditing: boolean;
  isLocked: boolean;
  isDirty: boolean;
  isLoading: boolean;
  saveStatus: SaveStatus;
  error: string | null;

  // Slash menu
  showSlashMenu: boolean;
  slashFilter: string;
  slashBlockId: string | null;

  // Floating toolbar
  showToolbar: boolean;
  toolbarPosition: { x: number; y: number };
  selectedText: { start: number; end: number } | null;

  // Actions
  loadDocument: (documentId: string) => Promise<void>;
  createDocument: (title: string, chatId?: string, topicId?: string, templateId?: string) => Promise<string | null>;
  setDocument: (doc: DocumentFull) => void;
  updateTitle: (title: string) => void;
  updateIcon: (icon: string) => void;
  addBlock: (type: BlockType, afterBlockId?: string | null) => void;
  updateBlock: (blockId: string, changes: Partial<Block>) => void;
  deleteBlock: (blockId: string) => void;
  duplicateBlock: (blockId: string) => void;
  moveBlock: (blockId: string, direction: 'up' | 'down') => void;
  reorderBlocks: (fromIndex: number, toIndex: number) => void;
  changeBlockType: (blockId: string, newType: BlockType) => void;
  setActiveBlock: (blockId: string | null) => void;
  toggleChecklist: (blockId: string) => void;

  // Slash menu
  openSlashMenu: (blockId: string) => void;
  closeSlashMenu: () => void;
  setSlashFilter: (filter: string) => void;

  // Toolbar
  openToolbar: (x: number, y: number, selection: { start: number; end: number }) => void;
  closeToolbar: () => void;
  applyFormat: (format: string) => void;

  // Save
  save: () => Promise<void>;
  scheduleSave: () => void;
  reset: () => void;
};

let saveTimer: ReturnType<typeof setTimeout> | null = null;

const generateTempId = () => `temp-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

export const useEditorStore = create<EditorState>()((set, get) => ({
  documentId: null,
  title: '',
  icon: '📄',
  blocks: [],
  activeBlockId: null,
  isEditing: false,
  isLocked: false,
  isDirty: false,
  isLoading: false,
  saveStatus: 'idle',
  error: null,

  showSlashMenu: false,
  slashFilter: '',
  slashBlockId: null,

  showToolbar: false,
  toolbarPosition: { x: 0, y: 0 },
  selectedText: null,

  loadDocument: async (documentId: string) => {
    set({ isLoading: true, error: null });
    try {
      const res = await documentsApi.getById(documentId);
      const doc = res.data.data;
      set({
        documentId: doc.document.id,
        title: doc.document.title,
        icon: doc.document.icon,
        blocks: doc.blocks ?? [],
        isLocked: doc.document.locked,
        isEditing: true,
        isLoading: false,
        isDirty: false,
        saveStatus: 'idle',
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : i18n.t('document.loadFailed');
      set({ error: msg, isLoading: false });
    }
  },

  createDocument: async (title, chatId, topicId, templateId) => {
    set({ isLoading: true, error: null });
    try {
      const res = await documentsApi.create({
        title,
        chatId,
        topicId,
        isStandalone: !chatId && !topicId,
        templateId,
      });
      const doc = res.data.data;
      set({
        documentId: doc.document.id,
        title: doc.document.title,
        icon: doc.document.icon,
        blocks: doc.blocks ?? [],
        isLocked: false,
        isEditing: true,
        isLoading: false,
        isDirty: false,
        saveStatus: 'idle',
      });
      return doc.document.id;
    } catch (err) {
      const msg = err instanceof Error ? err.message : i18n.t('document.createFailed');
      set({ error: msg, isLoading: false });
      return null;
    }
  },

  setDocument: (doc: DocumentFull) => {
    set({
      documentId: doc.document.id,
      title: doc.document.title,
      icon: doc.document.icon,
      blocks: doc.blocks ?? [],
      isLocked: doc.document.locked,
      isEditing: true,
      isDirty: false,
      saveStatus: 'idle',
    });
  },

  updateTitle: (title: string) => {
    set({ title, isDirty: true });
    get().scheduleSave();
  },

  updateIcon: (icon: string) => {
    set({ icon, isDirty: true });
    get().scheduleSave();
  },

  addBlock: (type: BlockType, afterBlockId?: string | null) => {
    set((state) => {
      const blocks = [...state.blocks];
      let insertIndex = blocks.length;

      if (afterBlockId) {
        const idx = blocks.findIndex((b) => b.id === afterBlockId);
        if (idx !== -1) insertIndex = idx + 1;
      }

      const newBlock: Block = {
        id: generateTempId(),
        documentId: state.documentId ?? '',
        type,
        content: '',
        sortOrder: insertIndex,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      blocks.splice(insertIndex, 0, newBlock);

      // Re-index sortOrder
      const reindexed: Block[] = blocks.map((b, i) => ({ ...b, sortOrder: i }));

      return {
        blocks: reindexed,
        activeBlockId: newBlock.id,
        isDirty: true,
        showSlashMenu: false,
        slashFilter: '',
      };
    });
    get().scheduleSave();
  },

  updateBlock: (blockId: string, changes: Partial<Block>) => {
    set((state) => ({
      blocks: state.blocks.map((b) =>
        b.id === blockId ? { ...b, ...changes, updatedAt: new Date().toISOString() } : b,
      ),
      isDirty: true,
    }));
    get().scheduleSave();
  },

  deleteBlock: (blockId: string) => {
    set((state) => {
      const blocks = state.blocks.filter((b) => b.id !== blockId);
      // If no blocks left, add empty paragraph
      if (blocks.length === 0) {
        blocks.push({
          id: generateTempId(),
          documentId: state.documentId ?? '',
          type: 'paragraph',
          content: '',
          sortOrder: 0,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });
      }
      const reindexed: Block[] = blocks.map((b, i) => ({ ...b, sortOrder: i }));
      // Move active to previous block
      const deletedIdx = state.blocks.findIndex((b) => b.id === blockId);
      const newActiveIdx = Math.max(0, deletedIdx - 1);

      return {
        blocks: reindexed,
        activeBlockId: reindexed[newActiveIdx]?.id ?? null,
        isDirty: true,
      };
    });
    get().scheduleSave();
  },

  duplicateBlock: (blockId: string) => {
    set((state) => {
      const idx = state.blocks.findIndex((b) => b.id === blockId);
      if (idx === -1) return state;

      const original = state.blocks[idx];
      if (!original) return state;
      const copy: Block = {
        ...original,
        id: generateTempId(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      const blocks = [...state.blocks];
      blocks.splice(idx + 1, 0, copy);
      const reindexed: Block[] = blocks.map((b, i) => ({ ...b, sortOrder: i }));

      return { blocks: reindexed, isDirty: true, activeBlockId: copy.id };
    });
    get().scheduleSave();
  },

  moveBlock: (blockId: string, direction: 'up' | 'down') => {
    set((state) => {
      const blocks = [...state.blocks];
      const idx = blocks.findIndex((b) => b.id === blockId);
      if (idx === -1) return state;

      const targetIdx = direction === 'up' ? idx - 1 : idx + 1;
      if (targetIdx < 0 || targetIdx >= blocks.length) return state;

      const temp = blocks[idx]!;
      blocks[idx] = blocks[targetIdx]!;
      blocks[targetIdx] = temp;
      const reindexed: Block[] = blocks.map((b, i) => ({ ...b, sortOrder: i }));

      return { blocks: reindexed, isDirty: true };
    });
    get().scheduleSave();
  },

  reorderBlocks: (fromIndex: number, toIndex: number) => {
    set((state) => {
      const blocks = [...state.blocks];
      const [moved] = blocks.splice(fromIndex, 1);
      if (!moved) return state;
      blocks.splice(toIndex, 0, moved);
      const reindexed: Block[] = blocks.map((b, i) => ({ ...b, sortOrder: i }));
      return { blocks: reindexed, isDirty: true };
    });
    get().scheduleSave();
  },

  changeBlockType: (blockId: string, newType: BlockType) => {
    set((state) => ({
      blocks: state.blocks.map((b) =>
        b.id === blockId
          ? { ...b, type: newType, updatedAt: new Date().toISOString() }
          : b,
      ),
      isDirty: true,
      showSlashMenu: false,
      slashFilter: '',
    }));
    get().scheduleSave();
  },

  setActiveBlock: (blockId: string | null) => {
    set({ activeBlockId: blockId });
  },

  toggleChecklist: (blockId: string) => {
    set((state) => ({
      blocks: state.blocks.map((b) =>
        b.id === blockId
          ? { ...b, checked: !b.checked, updatedAt: new Date().toISOString() }
          : b,
      ),
      isDirty: true,
    }));
    get().scheduleSave();
  },

  // Slash menu
  openSlashMenu: (blockId: string) => {
    set({ showSlashMenu: true, slashBlockId: blockId, slashFilter: '' });
  },

  closeSlashMenu: () => {
    set({ showSlashMenu: false, slashBlockId: null, slashFilter: '' });
  },

  setSlashFilter: (filter: string) => {
    set({ slashFilter: filter });
  },

  // Toolbar
  openToolbar: (x: number, y: number, selection: { start: number; end: number }) => {
    set({ showToolbar: true, toolbarPosition: { x, y }, selectedText: selection });
  },

  closeToolbar: () => {
    set({ showToolbar: false, selectedText: null });
  },

  applyFormat: (format: string) => {
    const state = get();
    if (!state.activeBlockId || !state.selectedText) return;

    const block = state.blocks.find((b) => b.id === state.activeBlockId);
    if (!block) return;

    const { start, end } = state.selectedText;
    const text = block.content;
    const selected = text.slice(start, end);

    let wrapper = '';
    switch (format) {
      case 'bold': wrapper = '**'; break;
      case 'italic': wrapper = '*'; break;
      case 'underline': wrapper = '__'; break;
      case 'strikethrough': wrapper = '~~'; break;
      case 'code': wrapper = '`'; break;
      default: return;
    }

    const newContent = text.slice(0, start) + wrapper + selected + wrapper + text.slice(end);

    set((s) => ({
      blocks: s.blocks.map((b) =>
        b.id === s.activeBlockId
          ? { ...b, content: newContent, updatedAt: new Date().toISOString() }
          : b,
      ),
      isDirty: true,
      showToolbar: false,
      selectedText: null,
    }));
    get().scheduleSave();
  },

  // Save
  save: async () => {
    const state = get();
    if (!state.documentId || !state.isDirty) return;

    set({ saveStatus: 'saving' });
    try {
      // Update document title/icon
      await documentsApi.update(state.documentId, {
        title: state.title,
        icon: state.icon,
      });

      // Batch update blocks
      const operations = state.blocks.map((b) => ({
        action: b.id.startsWith('temp-') ? 'add' : 'update',
        block: {
          id: b.id.startsWith('temp-') ? undefined : b.id,
          type: b.type,
          content: b.content,
          checked: b.checked,
          rows: b.rows,
          columns: b.columns,
          language: b.language,
          emoji: b.emoji,
          color: b.color,
          sortOrder: b.sortOrder,
        },
      }));

      await documentsApi.batchBlocks(state.documentId, operations);

      // Reload to get real block IDs
      const res = await documentsApi.getById(state.documentId);
      const doc = res.data.data;

      set({
        blocks: doc.blocks ?? [],
        isDirty: false,
        saveStatus: 'saved',
      });

      // Reset status after a moment
      setTimeout(() => {
        if (get().saveStatus === 'saved') {
          set({ saveStatus: 'idle' });
        }
      }, 2000);
    } catch {
      set({ saveStatus: 'error' });
      // Retry after 5s
      setTimeout(() => {
        get().save();
      }, 5000);
    }
  },

  scheduleSave: () => {
    if (saveTimer) clearTimeout(saveTimer);
    saveTimer = setTimeout(() => {
      get().save();
    }, 2000);
  },

  reset: () => {
    if (saveTimer) clearTimeout(saveTimer);
    set({
      documentId: null,
      title: '',
      icon: '📄',
      blocks: [],
      activeBlockId: null,
      isEditing: false,
      isLocked: false,
      isDirty: false,
      isLoading: false,
      saveStatus: 'idle',
      error: null,
      showSlashMenu: false,
      slashFilter: '',
      slashBlockId: null,
      showToolbar: false,
      toolbarPosition: { x: 0, y: 0 },
      selectedText: null,
    });
  },
}));
