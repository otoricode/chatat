// CRDT client for real-time collaborative editing
// Uses LWW (Last-Writer-Wins) Register with Lamport timestamps

/**
 * Lamport clock for generating monotonically increasing timestamps
 * with hybrid logical clock behavior.
 */
export class LamportClock {
  private counter: number;
  readonly nodeId: string;

  constructor(nodeId: string) {
    this.counter = Date.now();
    this.nodeId = nodeId;
  }

  /** Advance the clock and return a new timestamp. */
  tick(): number {
    const now = Date.now();
    if (now > this.counter) {
      this.counter = now;
    } else {
      this.counter++;
    }
    return this.counter;
  }

  /** Update clock based on a received remote timestamp. */
  receiveTick(remoteTs: number): number {
    if (remoteTs > this.counter) {
      this.counter = remoteTs;
    }
    this.counter++;
    return this.counter;
  }

  get current(): number {
    return this.counter;
  }
}

/** A single LWW register value. */
export interface LWWValue {
  value: string;
  timestamp: number;
  nodeId: string;
}

/**
 * Merges a remote LWW value into a local one.
 * Returns true if the remote value wins (was accepted).
 */
export function mergeLWW(local: LWWValue, remote: LWWValue): boolean {
  if (remote.timestamp > local.timestamp) {
    local.value = remote.value;
    local.timestamp = remote.timestamp;
    local.nodeId = remote.nodeId;
    return true;
  }
  if (remote.timestamp === local.timestamp && remote.nodeId > local.nodeId) {
    local.value = remote.value;
    local.timestamp = remote.timestamp;
    local.nodeId = remote.nodeId;
    return true;
  }
  return false;
}

/** CRDT update event sent/received over WebSocket. */
export interface CRDTUpdateEvent {
  documentId: string;
  blockId: string;
  field: string; // "content" | "checked"
  value: string;
  timestamp: number;
  nodeId: string;
  action?: 'update' | 'delete';
}

/**
 * Per-block CRDT state tracking.
 * Tracks the latest known value for content and checked fields.
 */
export interface BlockCRDTState {
  blockId: string;
  content: LWWValue;
  checked: LWWValue;
  deleted: boolean;
  deletedAt: number;
}

function newBlockState(blockId: string): BlockCRDTState {
  return {
    blockId,
    content: { value: '', timestamp: 0, nodeId: '' },
    checked: { value: '', timestamp: 0, nodeId: '' },
    deleted: false,
    deletedAt: 0,
  };
}

/**
 * DocumentCRDT manages CRDT state for all blocks in a document.
 * It merges incoming remote updates and determines if they should
 * be applied to the local editor state.
 */
export class DocumentCRDT {
  readonly documentId: string;
  readonly clock: LamportClock;
  private blocks = new Map<string, BlockCRDTState>();

  constructor(documentId: string, nodeId: string) {
    this.documentId = documentId;
    this.clock = new LamportClock(nodeId);
  }

  /**
   * Apply a remote update event.
   * Returns true if the remote value was accepted (wins over local).
   */
  applyRemoteUpdate(event: CRDTUpdateEvent): boolean {
    this.clock.receiveTick(event.timestamp);

    let state = this.blocks.get(event.blockId);
    if (!state) {
      state = newBlockState(event.blockId);
      this.blocks.set(event.blockId, state);
    }

    if (state.deleted) return false;

    const remote: LWWValue = {
      value: event.value,
      timestamp: event.timestamp,
      nodeId: event.nodeId,
    };

    switch (event.field) {
      case 'content':
        return mergeLWW(state.content, remote);
      case 'checked':
        return mergeLWW(state.checked, remote);
      default:
        return false;
    }
  }

  /** Apply a remote delete event. Returns true if accepted. */
  applyRemoteDelete(event: CRDTUpdateEvent): boolean {
    this.clock.receiveTick(event.timestamp);

    let state = this.blocks.get(event.blockId);
    if (!state) {
      state = newBlockState(event.blockId);
      this.blocks.set(event.blockId, state);
    }

    if (state.deleted && event.timestamp <= state.deletedAt) {
      return false;
    }

    state.deleted = true;
    state.deletedAt = event.timestamp;
    return true;
  }

  /**
   * Create a local update event for sending to the server.
   * Advances the Lamport clock and returns the event.
   */
  createUpdateEvent(blockId: string, field: string, value: string): CRDTUpdateEvent {
    const timestamp = this.clock.tick();

    // Track locally
    let state = this.blocks.get(blockId);
    if (!state) {
      state = newBlockState(blockId);
      this.blocks.set(blockId, state);
    }

    const lwv: LWWValue = { value, timestamp, nodeId: this.clock.nodeId };
    if (field === 'content') {
      state.content = lwv;
    } else if (field === 'checked') {
      state.checked = lwv;
    }

    return {
      documentId: this.documentId,
      blockId,
      field,
      value,
      timestamp,
      nodeId: this.clock.nodeId,
      action: 'update',
    };
  }

  /** Create a local delete event for sending to the server. */
  createDeleteEvent(blockId: string): CRDTUpdateEvent {
    const timestamp = this.clock.tick();

    let state = this.blocks.get(blockId);
    if (!state) {
      state = newBlockState(blockId);
      this.blocks.set(blockId, state);
    }
    state.deleted = true;
    state.deletedAt = timestamp;

    return {
      documentId: this.documentId,
      blockId,
      field: '',
      value: '',
      timestamp,
      nodeId: this.clock.nodeId,
      action: 'delete',
    };
  }

  /** Get block CRDT state if it exists. */
  getBlockState(blockId: string): BlockCRDTState | undefined {
    return this.blocks.get(blockId);
  }

  /** Clear all state (e.g., when leaving the document). */
  clear(): void {
    this.blocks.clear();
  }
}
