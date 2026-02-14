// @ts-nocheck
jest.mock('@/services/api/documents', () => ({
  documentsApi: {
    getById: jest.fn(),
    create: jest.fn(),
    update: jest.fn(),
    batchBlocks: jest.fn(),
  },
}));

import { useEditorStore } from '../editorStore';
import { documentsApi } from '@/services/api/documents';

const mockDocsApi = documentsApi as jest.Mocked<typeof documentsApi>;

const makeBlock = (id: string, type: string, content: string, sortOrder = 0) => ({
  id,
  documentId: 'doc1',
  type,
  content,
  sortOrder,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
});

const makeDocFull = (id: string, title: string, blocks: any[] = []) => ({
  document: { id, title, icon: '📄', locked: false, ownerId: 'u1', contextType: 'standalone' },
  blocks,
  collaborators: [],
  signers: [],
});

beforeEach(() => {
  useEditorStore.getState().reset();
  jest.clearAllMocks();
  jest.useFakeTimers();
});

afterEach(() => {
  jest.useRealTimers();
});

describe('editorStore', () => {
  it('starts with default state', () => {
    const s = useEditorStore.getState();
    expect(s.documentId).toBeNull();
    expect(s.title).toBe('');
    expect(s.blocks).toEqual([]);
    expect(s.isEditing).toBe(false);
    expect(s.isDirty).toBe(false);
    expect(s.saveStatus).toBe('idle');
  });

  it('loadDocument success', async () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello')];
    mockDocsApi.getById.mockResolvedValue({
      data: { data: makeDocFull('doc1', 'Test Doc', blocks) },
    } as any);

    await useEditorStore.getState().loadDocument('doc1');

    const s = useEditorStore.getState();
    expect(s.documentId).toBe('doc1');
    expect(s.title).toBe('Test Doc');
    expect(s.blocks).toHaveLength(1);
    expect(s.isEditing).toBe(true);
    expect(s.isLoading).toBe(false);
  });

  it('loadDocument handles null blocks', async () => {
    mockDocsApi.getById.mockResolvedValue({
      data: { data: makeDocFull('doc1', 'Test', null as any) },
    } as any);

    await useEditorStore.getState().loadDocument('doc1');

    expect(useEditorStore.getState().blocks).toEqual([]);
  });

  it('loadDocument error with Error', async () => {
    mockDocsApi.getById.mockRejectedValue(new Error('Network fail'));

    await useEditorStore.getState().loadDocument('doc1');

    expect(useEditorStore.getState().error).toBe('Network fail');
    expect(useEditorStore.getState().isLoading).toBe(false);
  });

  it('loadDocument error with non-Error', async () => {
    mockDocsApi.getById.mockRejectedValue('err');

    await useEditorStore.getState().loadDocument('doc1');

    // i18n mock returns key
    expect(useEditorStore.getState().error).toBe('document.loadFailed');
  });

  it('createDocument success', async () => {
    const doc = makeDocFull('doc1', 'New Doc');
    mockDocsApi.create.mockResolvedValue({
      data: { data: doc },
    } as any);

    const id = await useEditorStore.getState().createDocument('New Doc');

    expect(id).toBe('doc1');
    expect(useEditorStore.getState().documentId).toBe('doc1');
    expect(useEditorStore.getState().isEditing).toBe(true);
  });

  it('createDocument error returns null', async () => {
    mockDocsApi.create.mockRejectedValue(new Error('Create failed'));

    const id = await useEditorStore.getState().createDocument('New');

    expect(id).toBeNull();
    expect(useEditorStore.getState().error).toBe('Create failed');
  });

  it('createDocument error with non-Error', async () => {
    mockDocsApi.create.mockRejectedValue(42);

    const id = await useEditorStore.getState().createDocument('New');

    expect(id).toBeNull();
    expect(useEditorStore.getState().error).toBe('document.createFailed');
  });

  it('setDocument sets all fields', () => {
    const doc = makeDocFull('doc1', 'Set Doc', [makeBlock('b1', 'paragraph', 'Hi')]);

    useEditorStore.getState().setDocument(doc as any);

    const s = useEditorStore.getState();
    expect(s.documentId).toBe('doc1');
    expect(s.title).toBe('Set Doc');
    expect(s.blocks).toHaveLength(1);
    expect(s.isEditing).toBe(true);
    expect(s.isDirty).toBe(false);
  });

  it('updateTitle sets title and marks dirty', () => {
    useEditorStore.getState().updateTitle('New Title');

    const s = useEditorStore.getState();
    expect(s.title).toBe('New Title');
    expect(s.isDirty).toBe(true);
  });

  it('updateIcon sets icon and marks dirty', () => {
    useEditorStore.getState().updateIcon('🎉');

    const s = useEditorStore.getState();
    expect(s.icon).toBe('🎉');
    expect(s.isDirty).toBe(true);
  });

  it('addBlock appends to end', () => {
    useEditorStore.setState({ documentId: 'doc1', blocks: [] });

    useEditorStore.getState().addBlock('paragraph');

    const s = useEditorStore.getState();
    expect(s.blocks).toHaveLength(1);
    expect(s.blocks[0].type).toBe('paragraph');
    expect(s.isDirty).toBe(true);
  });

  it('addBlock inserts after specified block', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'First'), makeBlock('b2', 'paragraph', 'Second')];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().addBlock('heading_1', 'b1');

    const s = useEditorStore.getState();
    expect(s.blocks).toHaveLength(3);
    expect(s.blocks[1].type).toBe('heading_1');
    expect(s.blocks[0].sortOrder).toBe(0);
    expect(s.blocks[1].sortOrder).toBe(1);
    expect(s.blocks[2].sortOrder).toBe(2);
  });

  it('updateBlock updates specific block', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello')];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().updateBlock('b1', { content: 'Updated' });

    expect(useEditorStore.getState().blocks[0].content).toBe('Updated');
    expect(useEditorStore.getState().isDirty).toBe(true);
  });

  it('deleteBlock removes block', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'First'), makeBlock('b2', 'paragraph', 'Second')];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().deleteBlock('b1');

    const s = useEditorStore.getState();
    expect(s.blocks).toHaveLength(1);
    expect(s.blocks[0].id).toBe('b2');
    expect(s.isDirty).toBe(true);
  });

  it('deleteBlock adds empty paragraph when last block deleted', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Only')];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().deleteBlock('b1');

    const s = useEditorStore.getState();
    expect(s.blocks).toHaveLength(1);
    expect(s.blocks[0].type).toBe('paragraph');
    expect(s.blocks[0].content).toBe('');
  });

  it('duplicateBlock creates copy', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Original')];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().duplicateBlock('b1');

    const s = useEditorStore.getState();
    expect(s.blocks).toHaveLength(2);
    expect(s.blocks[1].content).toBe('Original');
    expect(s.blocks[1].id).not.toBe('b1');
    expect(s.isDirty).toBe(true);
  });

  it('duplicateBlock with nonexistent id does nothing', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Only')];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().duplicateBlock('nonexistent');

    expect(useEditorStore.getState().blocks).toHaveLength(1);
  });

  it('moveBlock up', () => {
    const blocks = [
      makeBlock('b1', 'paragraph', 'First', 0),
      makeBlock('b2', 'paragraph', 'Second', 1),
    ];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().moveBlock('b2', 'up');

    const s = useEditorStore.getState();
    expect(s.blocks[0].id).toBe('b2');
    expect(s.blocks[1].id).toBe('b1');
  });

  it('moveBlock down', () => {
    const blocks = [
      makeBlock('b1', 'paragraph', 'First', 0),
      makeBlock('b2', 'paragraph', 'Second', 1),
    ];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().moveBlock('b1', 'down');

    const s = useEditorStore.getState();
    expect(s.blocks[0].id).toBe('b2');
    expect(s.blocks[1].id).toBe('b1');
  });

  it('moveBlock up at boundary does nothing', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'First', 0)];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().moveBlock('b1', 'up');

    expect(useEditorStore.getState().blocks[0].id).toBe('b1');
  });

  it('moveBlock with invalid id does nothing', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'First', 0)];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().moveBlock('nonexistent', 'up');

    expect(useEditorStore.getState().blocks).toHaveLength(1);
  });

  it('reorderBlocks swaps positions', () => {
    const blocks = [
      makeBlock('b1', 'paragraph', 'First', 0),
      makeBlock('b2', 'paragraph', 'Second', 1),
      makeBlock('b3', 'paragraph', 'Third', 2),
    ];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().reorderBlocks(0, 2);

    const s = useEditorStore.getState();
    expect(s.blocks[0].id).toBe('b2');
    expect(s.blocks[1].id).toBe('b3');
    expect(s.blocks[2].id).toBe('b1');
  });

  it('changeBlockType updates type', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello')];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().changeBlockType('b1', 'heading_1');

    expect(useEditorStore.getState().blocks[0].type).toBe('heading_1');
    expect(useEditorStore.getState().isDirty).toBe(true);
    expect(useEditorStore.getState().showSlashMenu).toBe(false);
  });

  it('setActiveBlock sets activeBlockId', () => {
    useEditorStore.getState().setActiveBlock('b1');

    expect(useEditorStore.getState().activeBlockId).toBe('b1');
  });

  it('setActiveBlock with null clears it', () => {
    useEditorStore.setState({ activeBlockId: 'b1' });

    useEditorStore.getState().setActiveBlock(null);

    expect(useEditorStore.getState().activeBlockId).toBeNull();
  });

  it('toggleChecklist toggles checked', () => {
    const blocks = [{ ...makeBlock('b1', 'checklist', 'Task'), checked: false }];
    useEditorStore.setState({ documentId: 'doc1', blocks: blocks as any });

    useEditorStore.getState().toggleChecklist('b1');

    expect((useEditorStore.getState().blocks[0] as any).checked).toBe(true);
  });

  it('openSlashMenu sets state', () => {
    useEditorStore.getState().openSlashMenu('b1');

    const s = useEditorStore.getState();
    expect(s.showSlashMenu).toBe(true);
    expect(s.slashBlockId).toBe('b1');
    expect(s.slashFilter).toBe('');
  });

  it('closeSlashMenu clears state', () => {
    useEditorStore.setState({ showSlashMenu: true, slashBlockId: 'b1', slashFilter: 'he' });

    useEditorStore.getState().closeSlashMenu();

    const s = useEditorStore.getState();
    expect(s.showSlashMenu).toBe(false);
    expect(s.slashBlockId).toBeNull();
    expect(s.slashFilter).toBe('');
  });

  it('setSlashFilter updates filter', () => {
    useEditorStore.getState().setSlashFilter('heading');

    expect(useEditorStore.getState().slashFilter).toBe('heading');
  });

  it('openToolbar sets position and selection', () => {
    useEditorStore.getState().openToolbar(100, 200, { start: 5, end: 10 });

    const s = useEditorStore.getState();
    expect(s.showToolbar).toBe(true);
    expect(s.toolbarPosition).toEqual({ x: 100, y: 200 });
    expect(s.selectedText).toEqual({ start: 5, end: 10 });
  });

  it('closeToolbar clears state', () => {
    useEditorStore.setState({ showToolbar: true, selectedText: { start: 0, end: 5 } });

    useEditorStore.getState().closeToolbar();

    expect(useEditorStore.getState().showToolbar).toBe(false);
    expect(useEditorStore.getState().selectedText).toBeNull();
  });

  it('applyFormat bold wraps selection', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello World')];
    useEditorStore.setState({
      documentId: 'doc1',
      blocks: blocks as any,
      activeBlockId: 'b1',
      selectedText: { start: 6, end: 11 },
    });

    useEditorStore.getState().applyFormat('bold');

    expect(useEditorStore.getState().blocks[0].content).toBe('Hello **World**');
    expect(useEditorStore.getState().showToolbar).toBe(false);
  });

  it('applyFormat italic wraps selection', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello World')];
    useEditorStore.setState({
      documentId: 'doc1',
      blocks: blocks as any,
      activeBlockId: 'b1',
      selectedText: { start: 0, end: 5 },
    });

    useEditorStore.getState().applyFormat('italic');

    expect(useEditorStore.getState().blocks[0].content).toBe('*Hello* World');
  });

  it('applyFormat underline wraps selection', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello World')];
    useEditorStore.setState({
      documentId: 'doc1',
      blocks: blocks as any,
      activeBlockId: 'b1',
      selectedText: { start: 0, end: 5 },
    });

    useEditorStore.getState().applyFormat('underline');

    expect(useEditorStore.getState().blocks[0].content).toBe('__Hello__ World');
  });

  it('applyFormat strikethrough wraps selection', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello World')];
    useEditorStore.setState({
      documentId: 'doc1',
      blocks: blocks as any,
      activeBlockId: 'b1',
      selectedText: { start: 0, end: 5 },
    });

    useEditorStore.getState().applyFormat('strikethrough');

    expect(useEditorStore.getState().blocks[0].content).toBe('~~Hello~~ World');
  });

  it('applyFormat code wraps selection', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello World')];
    useEditorStore.setState({
      documentId: 'doc1',
      blocks: blocks as any,
      activeBlockId: 'b1',
      selectedText: { start: 0, end: 5 },
    });

    useEditorStore.getState().applyFormat('code');

    expect(useEditorStore.getState().blocks[0].content).toBe('`Hello` World');
  });

  it('applyFormat unknown format does nothing', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello World')];
    useEditorStore.setState({
      documentId: 'doc1',
      blocks: blocks as any,
      activeBlockId: 'b1',
      selectedText: { start: 0, end: 5 },
    });

    useEditorStore.getState().applyFormat('unknown');

    expect(useEditorStore.getState().blocks[0].content).toBe('Hello World');
  });

  it('applyFormat without activeBlockId does nothing', () => {
    useEditorStore.setState({
      activeBlockId: null,
      selectedText: { start: 0, end: 5 },
    });

    useEditorStore.getState().applyFormat('bold');

    // No crash
  });

  it('applyFormat without selectedText does nothing', () => {
    useEditorStore.setState({
      activeBlockId: 'b1',
      selectedText: null,
    });

    useEditorStore.getState().applyFormat('bold');

    // No crash
  });

  it('applyFormat with non-matching block does nothing', () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello')];
    useEditorStore.setState({
      documentId: 'doc1',
      blocks: blocks as any,
      activeBlockId: 'nonexistent',
      selectedText: { start: 0, end: 5 },
    });

    useEditorStore.getState().applyFormat('bold');

    expect(useEditorStore.getState().blocks[0].content).toBe('Hello');
  });

  it('save skips when no documentId', async () => {
    useEditorStore.setState({ documentId: null, isDirty: true });

    await useEditorStore.getState().save();

    expect(mockDocsApi.update).not.toHaveBeenCalled();
  });

  it('save skips when not dirty', async () => {
    useEditorStore.setState({ documentId: 'doc1', isDirty: false });

    await useEditorStore.getState().save();

    expect(mockDocsApi.update).not.toHaveBeenCalled();
  });

  it('save success', async () => {
    const blocks = [makeBlock('b1', 'paragraph', 'Hello')];
    useEditorStore.setState({
      documentId: 'doc1',
      title: 'Test',
      icon: '📄',
      blocks: blocks as any,
      isDirty: true,
    });

    mockDocsApi.update.mockResolvedValue({} as any);
    mockDocsApi.batchBlocks.mockResolvedValue({} as any);
    mockDocsApi.getById.mockResolvedValue({
      data: { data: makeDocFull('doc1', 'Test', blocks) },
    } as any);

    await useEditorStore.getState().save();

    const s = useEditorStore.getState();
    expect(s.saveStatus).toBe('saved');
    expect(s.isDirty).toBe(false);
  });

  it('save error sets error status', async () => {
    useEditorStore.setState({
      documentId: 'doc1',
      title: 'Test',
      blocks: [makeBlock('b1', 'paragraph', 'Hello')] as any,
      isDirty: true,
    });

    mockDocsApi.update.mockRejectedValue(new Error('Save failed'));

    await useEditorStore.getState().save();

    expect(useEditorStore.getState().saveStatus).toBe('error');
  });

  it('reset clears all state', () => {
    useEditorStore.setState({
      documentId: 'doc1',
      title: 'Test',
      blocks: [makeBlock('b1', 'paragraph', 'Hi')] as any,
      isDirty: true,
      isEditing: true,
    });

    useEditorStore.getState().reset();

    const s = useEditorStore.getState();
    expect(s.documentId).toBeNull();
    expect(s.title).toBe('');
    expect(s.blocks).toEqual([]);
    expect(s.isDirty).toBe(false);
    expect(s.isEditing).toBe(false);
  });
});
