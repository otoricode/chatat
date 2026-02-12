# Phase 13: Block Editor (Mobile)

> Implementasi Notion-style block editor di React Native.
> Editor mendukung 13 tipe block, slash commands, floating toolbar.

**Estimasi:** 5 hari
**Dependency:** Phase 06 (Mobile Shell), Phase 12 (Document Data Layer)
**Output:** Rich block editor component yang production-ready.

---

## Task 13.1: Core Editor Architecture

**Input:** Phase 06 components, Phase 12 API
**Output:** Block editor state management dan rendering pipeline

### Steps:
1. Buat editor state store (`src/stores/editorStore.ts`):
   ```typescript
   interface EditorState {
     documentId: string | null;
     blocks: Block[];
     activeBlockId: string | null;
     isEditing: boolean;
     isLocked: boolean;
     isDirty: boolean;

     // Actions
     setDocument: (doc: DocumentFull) => void;
     addBlock: (type: BlockType, position: number) => void;
     updateBlock: (blockId: string, content: Partial<Block>) => void;
     deleteBlock: (blockId: string) => void;
     moveBlock: (blockId: string, newPosition: number) => void;
     setActiveBlock: (blockId: string | null) => void;
     save: () => Promise<void>;
   }

   type BlockType =
     | 'paragraph'
     | 'heading_1'
     | 'heading_2'
     | 'heading_3'
     | 'bullet_list'
     | 'numbered_list'
     | 'checklist'
     | 'table'
     | 'callout'
     | 'code'
     | 'toggle'
     | 'divider'
     | 'quote';
   ```
2. Buat `src/components/editor/BlockEditor.tsx`:
   ```typescript
   interface BlockEditorProps {
     documentId: string;
     readOnly?: boolean;
     onSave?: () => void;
   }

   const BlockEditor: React.FC<BlockEditorProps> = ({ documentId, readOnly }) => {
     // FlatList-based block rendering
     // Each block = separate component
     // Keyboard-aware scrolling
     // Auto-save on changes (debounced 2s)
   };
   ```
3. Block rendering strategy:
   - FlatList with `keyExtractor` by block.id
   - Memoized block components (`React.memo`)
   - Measure block heights for smooth scrolling
   - Keyboard offset management
4. Auto-save logic:
   - Debounce: 2 seconds after last change
   - Queue pending changes
   - Retry on network failure
   - Show save indicator (Menyimpan... / Tersimpan)

### Acceptance Criteria:
- [ ] Editor renders blocks from API data
- [ ] Zustand store manages editor state
- [ ] FlatList-based rendering (performant 100+ blocks)
- [ ] Auto-save debounced at 2s
- [ ] Save indicator shown
- [ ] Read-only mode for locked docs

### Testing:
- [ ] Unit test: editorStore actions (add, update, delete, move)
- [ ] Unit test: auto-save debounce logic
- [ ] Component test: BlockEditor renders blocks
- [ ] Component test: read-only mode disables editing

---

## Task 13.2: Text Block Components

**Input:** Task 13.1
**Output:** Paragraph, heading, quote, bullet, numbered, checklist blocks

### Steps:
1. Buat `src/components/editor/blocks/ParagraphBlock.tsx`:
   ```typescript
   const ParagraphBlock: React.FC<BlockProps> = ({ block, onChange, onFocus }) => {
     return (
       <TextInput
         style={styles.paragraph}
         value={block.content}
         onChangeText={(text) => onChange({ content: text })}
         onFocus={() => onFocus(block.id)}
         multiline
         placeholder="Ketik sesuatu..."
         placeholderTextColor="#6B7280"
       />
     );
   };

   const styles = StyleSheet.create({
     paragraph: {
       fontSize: 16,
       lineHeight: 24,
       color: '#E5E7EB',
       paddingVertical: 4,
       paddingHorizontal: 16,
     },
   });
   ```
2. Buat HeadingBlock (H1, H2, H3):
   - H1: fontSize 28, fontWeight 'bold'
   - H2: fontSize 22, fontWeight '600'
   - H3: fontSize 18, fontWeight '600'
3. Buat BulletListBlock:
   - Prefix: bullet dot (`•`)
   - Indent support (level 0, 1, 2)
   - Auto-create new bullet on Enter
   - Remove bullet on empty backspace
4. Buat NumberedListBlock:
   - Prefix: auto-numbered
   - Re-number on add/delete/reorder
   - Indent support
5. Buat ChecklistBlock:
   ```typescript
   const ChecklistBlock: React.FC<BlockProps> = ({ block, onChange }) => {
     return (
       <View style={styles.row}>
         <TouchableOpacity onPress={() => onChange({ checked: !block.checked })}>
           <View style={[styles.checkbox, block.checked && styles.checked]}>
             {block.checked && <CheckIcon size={14} color="#0F1117" />}
           </View>
         </TouchableOpacity>
         <TextInput
           style={[styles.text, block.checked && styles.strikethrough]}
           value={block.content}
           onChangeText={(text) => onChange({ content: text })}
         />
       </View>
     );
   };
   ```
6. Buat QuoteBlock:
   - Left border: 3px #6EE7B7
   - Italic text
   - Padding left: 16px

### Acceptance Criteria:
- [ ] Paragraph: editable, multiline, placeholder
- [ ] Headings: 3 levels, correct sizes
- [ ] Bullet list: auto-bullet, indent
- [ ] Numbered list: auto-number, re-number
- [ ] Checklist: toggle checked, strikethrough
- [ ] Quote: left accent border, italic
- [ ] Enter key: create new block of same type
- [ ] Backspace on empty: delete block, merge with previous

### Testing:
- [ ] Component test: each block type renders correctly
- [ ] Component test: text editing
- [ ] Component test: checklist toggle
- [ ] Component test: list auto-create on Enter
- [ ] Snapshot test: each block type

---

## Task 13.3: Special Block Components

**Input:** Task 13.1
**Output:** Table, callout, code, toggle, divider blocks

### Steps:
1. Buat `src/components/editor/blocks/TableBlock.tsx`:
   ```typescript
   interface TableData {
     headers: { name: string; type: 'text' | 'number' | 'date' | 'checkbox' }[];
     rows: string[][];
   }

   const TableBlock: React.FC<BlockProps> = ({ block, onChange }) => {
     const table: TableData = JSON.parse(block.rows || '{"headers":[],"rows":[]}');

     return (
       <ScrollView horizontal>
         <View>
           {/* Header row */}
           <View style={styles.headerRow}>
             {table.headers.map((h, i) => (
               <TextInput
                 key={`h-${i}`}
                 style={styles.headerCell}
                 value={h}
                 onChangeText={(text) => updateHeader(i, text)}
               />
             ))}
             <TouchableOpacity onPress={addColumn}>
               <PlusIcon />
             </TouchableOpacity>
           </View>
           {/* Data rows */}
           {table.rows.map((row, ri) => (
             <View key={`r-${ri}`} style={styles.dataRow}>
               {row.map((cell, ci) => (
                 <TextInput
                   key={`c-${ri}-${ci}`}
                   style={styles.dataCell}
                   value={cell}
                   onChangeText={(text) => updateCell(ri, ci, text)}
                 />
               ))}
             </View>
           ))}
           <TouchableOpacity onPress={addRow}>
             <Text style={styles.addRow}>+ Tambah Baris</Text>
           </TouchableOpacity>
         </View>
       </ScrollView>
     );
   };
   ```
2. Buat CalloutBlock:
   - Emoji selector (default: lamp bulb)
   - Background color options: blue, green, yellow, red, grey
   - Content area with text input
3. Buat CodeBlock:
   ```typescript
   const CodeBlock: React.FC<BlockProps> = ({ block, onChange }) => {
     return (
       <View style={styles.codeContainer}>
         <View style={styles.codeHeader}>
           <Text style={styles.language}>{block.language || 'text'}</Text>
           <TouchableOpacity onPress={copyCode}>
             <CopyIcon />
           </TouchableOpacity>
         </View>
         <TextInput
           style={styles.codeContent}
           value={block.content}
           onChangeText={(text) => onChange({ content: text })}
           multiline
           autoCapitalize="none"
           autoCorrect={false}
           fontFamily="monospace"
         />
       </View>
     );
   };
   ```
4. Buat ToggleBlock:
   - Expandable header (tap to toggle)
   - Children blocks inside (nested rendering)
   - Animated chevron icon
5. Buat DividerBlock:
   - Horizontal line
   - Non-editable
   - Height: 1px, color: #374151

### Acceptance Criteria:
- [ ] Table: add/remove rows & columns, edit cells, horizontal scroll
- [ ] Table: column type selector (Teks, Angka, Tanggal, Checkbox) saat buat kolom
- [ ] Table: drag-to-resize kolom (drag pembatas header kolom)
- [ ] Table: header row auto-styled
- [ ] Callout: emoji + color + content
- [ ] Code: monospace font, language label, copy button
- [ ] Toggle: expand/collapse, nested blocks
- [ ] Divider: renders horizontal line

### Testing:
- [ ] Component test: table add row/column
- [ ] Component test: callout with emoji
- [ ] Component test: code block copy
- [ ] Component test: toggle expand/collapse
- [ ] Snapshot test: all special blocks

---

## Task 13.4: Slash Command Menu

**Input:** Task 13.1, 13.2, 13.3
**Output:** Slash command (/) trigger dan block type selector

### Steps:
1. Buat `src/components/editor/SlashMenu.tsx`:
   ```typescript
   const BLOCK_OPTIONS: SlashOption[] = [
     { type: 'paragraph', icon: 'text', label: 'Teks', description: 'Teks biasa' },
     { type: 'heading_1', icon: 'heading-1', label: 'Judul 1', description: 'Judul besar' },
     { type: 'heading_2', icon: 'heading-2', label: 'Judul 2', description: 'Judul sedang' },
     { type: 'heading_3', icon: 'heading-3', label: 'Judul 3', description: 'Judul kecil' },
     { type: 'bullet_list', icon: 'list', label: 'Daftar Bullet', description: 'Daftar tak berurutan' },
     { type: 'numbered_list', icon: 'list-ordered', label: 'Daftar Nomor', description: 'Daftar berurutan' },
     { type: 'checklist', icon: 'check-square', label: 'Checklist', description: 'Daftar centang' },
     { type: 'table', icon: 'table', label: 'Tabel', description: 'Tabel data' },
     { type: 'callout', icon: 'alert-circle', label: 'Callout', description: 'Blok perhatian' },
     { type: 'code', icon: 'code', label: 'Kode', description: 'Blok kode' },
     { type: 'toggle', icon: 'chevron-right', label: 'Toggle', description: 'Blok toggle' },
     { type: 'divider', icon: 'minus', label: 'Pembatas', description: 'Garis horizontal' },
     { type: 'quote', icon: 'message-square', label: 'Kutipan', description: 'Blok kutipan' },
   ];

   const SlashMenu: React.FC<SlashMenuProps> = ({ position, onSelect, filter }) => {
     const filtered = BLOCK_OPTIONS.filter(
       (opt) => filter === '' || opt.label.toLowerCase().includes(filter.toLowerCase())
     );

     return (
       <Animated.View style={[styles.menu, { top: position.y }]}>
         <FlatList
           data={filtered}
           keyExtractor={(item) => item.type}
           renderItem={({ item }) => (
             <TouchableOpacity
               style={styles.option}
               onPress={() => onSelect(item.type)}
             >
               <Icon name={item.icon} size={20} color="#9CA3AF" />
               <View style={styles.optionText}>
                 <Text style={styles.label}>{item.label}</Text>
                 <Text style={styles.description}>{item.description}</Text>
               </View>
             </TouchableOpacity>
           )}
         />
       </Animated.View>
     );
   };
   ```
2. Trigger detection:
   - Detect "/" at block start
   - Show menu below cursor position
   - Filter options as user types after "/"
   - Keyboard navigation (up/down arrows)
   - Select: Enter or tap
   - Dismiss: Escape or tap outside
3. Block conversion:
   - On select → replace current block with new type
   - Remove "/" prefix from content
   - Focus new block immediately

### Acceptance Criteria:
- [ ] "/" triggers slash menu
- [ ] Menu shows all 13 block types
- [ ] Filter by typing after "/"
- [ ] Select replaces current block type
- [ ] Menu positioned below cursor
- [ ] Dismiss on Escape or outside tap
- [ ] Smooth animation (fade in/out)

### Testing:
- [ ] Component test: slash menu renders all options
- [ ] Component test: filter works
- [ ] Component test: select inserts block
- [ ] Component test: dismiss on outside tap
- [ ] Integration test: type "/" → select → block created

---

## Task 13.5: Floating Toolbar

**Input:** Task 13.2
**Output:** Context toolbar saat text diselect

### Steps:
1. Buat `src/components/editor/FloatingToolbar.tsx`:
   ```typescript
   const TOOLBAR_ACTIONS = [
     { id: 'bold', icon: 'bold', label: 'Bold' },
     { id: 'italic', icon: 'italic', label: 'Italic' },
     { id: 'underline', icon: 'underline', label: 'Underline' },
     { id: 'strikethrough', icon: 'strikethrough', label: 'Coret' },
     { id: 'code', icon: 'code', label: 'Kode' },
     { id: 'link', icon: 'link', label: 'Link' },
     { id: 'highlight', icon: 'highlight', label: 'Warna' },
   ];

   const FloatingToolbar: React.FC<ToolbarProps> = ({
     position,
     onAction,
     activeFormats,
   }) => {
     return (
       <Animated.View style={[styles.toolbar, { top: position.y, left: position.x }]}>
         {TOOLBAR_ACTIONS.map((action) => (
           <TouchableOpacity
             key={action.id}
             style={[styles.button, activeFormats.includes(action.id) && styles.active]}
             onPress={() => onAction(action.id)}
           >
             <Icon name={action.icon} size={18} color="#E5E7EB" />
           </TouchableOpacity>
         ))}
       </Animated.View>
     );
   };
   ```
2. Text selection detection:
   - onSelectionChange on TextInput
   - Show toolbar above selection
   - Position calculation (avoid screen edges)
3. Format application:
   - Apply markdown-style formatting to selection
   - Bold: `**text**`
   - Italic: `*text*`
   - Underline: `__text__`
   - Strikethrough: `~~text~~`
   - Inline code: `` `text` ``
   - Link: `[text](url)` → show URL input
   - Highlight: wrap with highlight marker, show color picker (4 colors: yellow, green, blue, pink)
4. Toolbar styling:
   - Dark bubble: background #1F2937
   - Rounded corners: borderRadius 8
   - Shadow for depth
   - Animated appearance

### Acceptance Criteria:
- [ ] Toolbar appears on text selection
- [ ] 7 formatting actions available (bold, italic, underline, strikethrough, code, link, highlight)
- [ ] Format applied to selected text
- [ ] Active formats highlighted
- [ ] Position: above selection, avoid edges
- [ ] Smooth animation

### Testing:
- [ ] Component test: toolbar renders all actions
- [ ] Component test: active format highlighting
- [ ] Unit test: format text (bold, italic, etc.)
- [ ] Integration test: select text → format → deselect

---

## Task 13.6: Block Actions & Drag-to-Reorder

**Input:** Task 13.1
**Output:** Block action menu dan drag-to-reorder

### Steps:
1. Block action menu (long press on block):
   ```typescript
   const BLOCK_ACTIONS = [
     { id: 'duplicate', icon: 'copy', label: 'Duplikat' },
     { id: 'delete', icon: 'trash-2', label: 'Hapus', destructive: true },
     { id: 'moveUp', icon: 'arrow-up', label: 'Pindah Atas' },
     { id: 'moveDown', icon: 'arrow-down', label: 'Pindah Bawah' },
     { id: 'changeType', icon: 'repeat', label: 'Ubah Tipe' },
   ];
   ```
2. Long press → show action menu (BottomSheet or popover)
3. Drag-to-reorder:
   - Drag handle on left side (6-dot grid icon)
   - react-native-reanimated for smooth drag
   - Haptic feedback on drag start
   - Visual placeholder during drag
   - Drop animation
4. Block type conversion:
   - "Ubah Tipe" → show SlashMenu-style selector
   - Convert content where possible (text preserved)
   - Cannot convert between incompatible types (table → paragraph)

### Acceptance Criteria:
- [ ] Long press shows action menu
- [ ] Duplicate, delete, move up/down work
- [ ] Block type conversion
- [ ] Drag handle visible on active block
- [ ] Smooth drag-and-drop reorder
- [ ] Haptic feedback on drag
- [ ] Incompatible conversion rejected gracefully

### Testing:
- [ ] Component test: action menu renders
- [ ] Unit test: block duplicate
- [ ] Unit test: block type conversion
- [ ] Integration test: drag reorder updates positions

---

## Phase 13 Review

### Testing Checklist:
- [ ] All 13 block types render correctly
- [ ] Slash command menu with filter
- [ ] Floating toolbar with formatting
- [ ] Block actions (duplicate, delete, move, change type)
- [ ] Drag-to-reorder
- [ ] Auto-save
- [ ] Read-only mode
- [ ] Performance: 100+ blocks smooth scroll
- [ ] Keyboard handling: no overlap, smooth scroll

### Review Checklist:
- [ ] Block types match `spesifikasi-chatat.md` section 5.3
- [ ] Labels in Indonesian
- [ ] Dark theme applied (#0F1117 bg, #6EE7B7 accent)
- [ ] Responsive on different screen sizes
- [ ] Consistent with WhatsApp-style design system
- [ ] Commit: `feat(editor): implement Notion-style block editor`
