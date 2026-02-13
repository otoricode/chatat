// useCollaborativeEditing — CRDT-based real-time collaboration hook
import { useEffect, useRef, useCallback } from 'react';
import { DocumentCRDT, type CRDTUpdateEvent } from '@/lib/crdt';
import { useEditorStore } from '@/stores/editorStore';
import { wsClient } from '@/services/ws';

// Generate a stable node ID for this device/session
const SESSION_NODE_ID = `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

type PresenceUser = {
  userId: string;
  action: 'joined' | 'left';
};

/**
 * Hook that manages CRDT-based collaborative editing for a document.
 * - Joins the document room on mount, leaves on unmount
 * - Sends CRDT update events when blocks are edited locally
 * - Receives remote CRDT updates and merges into editor state
 * - Tracks presence of other editors
 */
export function useCollaborativeEditing(documentId: string | null) {
  const crdtRef = useRef<DocumentCRDT | null>(null);
  const isRemoteUpdate = useRef(false);

  // Initialize CRDT when document changes
  useEffect(() => {
    if (!documentId) {
      crdtRef.current = null;
      return;
    }

    crdtRef.current = new DocumentCRDT(documentId, SESSION_NODE_ID);

    // Join the document room
    wsClient.send('doc_join', { documentId });

    // Listen for remote doc_update events
    const unsubUpdate = wsClient.on('doc_update', (payload: unknown) => {
      const event = payload as CRDTUpdateEvent;
      if (!event || event.documentId !== documentId) return;
      if (event.nodeId === SESSION_NODE_ID) return; // Ignore own echoes

      const crdt = crdtRef.current;
      if (!crdt) return;

      if (event.action === 'delete') {
        const accepted = crdt.applyRemoteDelete(event);
        if (accepted) {
          isRemoteUpdate.current = true;
          useEditorStore.getState().deleteBlock(event.blockId);
          isRemoteUpdate.current = false;
        }
      } else {
        const accepted = crdt.applyRemoteUpdate(event);
        if (accepted) {
          isRemoteUpdate.current = true;
          if (event.field === 'content') {
            useEditorStore.getState().updateBlock(event.blockId, { content: event.value });
          } else if (event.field === 'checked') {
            useEditorStore.getState().updateBlock(event.blockId, { checked: event.value === 'true' });
          }
          isRemoteUpdate.current = false;
        }
      }
    });

    // Listen for doc_lock events
    const unsubLock = wsClient.on('doc_lock', (payload: unknown) => {
      const data = payload as { documentId: string; locked: boolean };
      if (data.documentId !== documentId) return;
      useEditorStore.setState({ isLocked: data.locked });
    });

    return () => {
      // Leave document room
      wsClient.send('doc_leave', { documentId });
      crdtRef.current?.clear();
      crdtRef.current = null;
      unsubUpdate();
      unsubLock();
    };
  }, [documentId]);

  /**
   * Send a content update via CRDT.
   * Call this from the editor when a block's content changes locally.
   */
  const sendContentUpdate = useCallback(
    (blockId: string, content: string) => {
      if (isRemoteUpdate.current) return; // don't echo remote updates
      const crdt = crdtRef.current;
      if (!crdt) return;

      const event = crdt.createUpdateEvent(blockId, 'content', content);
      wsClient.send('doc_update', event);
    },
    [],
  );

  /**
   * Send a checked state update via CRDT.
   */
  const sendCheckedUpdate = useCallback(
    (blockId: string, checked: boolean) => {
      if (isRemoteUpdate.current) return;
      const crdt = crdtRef.current;
      if (!crdt) return;

      const event = crdt.createUpdateEvent(blockId, 'checked', String(checked));
      wsClient.send('doc_update', event);
    },
    [],
  );

  /**
   * Send a block delete event via CRDT.
   */
  const sendDeleteEvent = useCallback(
    (blockId: string) => {
      if (isRemoteUpdate.current) return;
      const crdt = crdtRef.current;
      if (!crdt) return;

      const event = crdt.createDeleteEvent(blockId);
      wsClient.send('doc_update', event);
    },
    [],
  );

  /** Whether the current update is from a remote peer (skip re-sending). */
  const isRemote = useCallback(() => isRemoteUpdate.current, []);

  return {
    sendContentUpdate,
    sendCheckedUpdate,
    sendDeleteEvent,
    isRemote,
  };
}

/**
 * Hook to track presence of other editors in a document room.
 */
export function useDocPresence(
  documentId: string | null,
  onPresenceChange?: (user: PresenceUser) => void,
) {
  const onPresenceRef = useRef(onPresenceChange);
  onPresenceRef.current = onPresenceChange;

  useEffect(() => {
    if (!documentId) return;

    const unsub = wsClient.on('doc_presence', (payload: unknown) => {
      const data = payload as PresenceUser & { documentId: string };
      if (data.documentId !== documentId) return;
      onPresenceRef.current?.({ userId: data.userId, action: data.action });
    });

    return unsub;
  }, [documentId]);
}
