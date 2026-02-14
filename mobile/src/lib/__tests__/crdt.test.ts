// @ts-nocheck
import {
  LamportClock,
  mergeLWW,
  DocumentCRDT,
  type LWWValue,
  type CRDTUpdateEvent,
} from '../crdt';

describe('LamportClock', () => {
  it('tick advances clock', () => {
    const clock = new LamportClock('node-1');
    const ts1 = clock.tick();
    const ts2 = clock.tick();
    expect(ts2).toBeGreaterThan(ts1);
  });

  it('tick increments when clock is ahead of Date.now', () => {
    const clock = new LamportClock('node-1');
    // Set clock far in the future
    (clock as any).counter = 99999999999999;
    const prev = clock.current;
    const ts = clock.tick();
    expect(ts).toBe(prev + 1);
  });

  it('receiveTick updates to remote if higher', () => {
    const clock = new LamportClock('node-1');
    const remoteTs = 99999999999999;
    const result = clock.receiveTick(remoteTs);
    expect(result).toBe(remoteTs + 1);
  });

  it('receiveTick increments if local is higher', () => {
    const clock = new LamportClock('node-1');
    (clock as any).counter = 99999999999999;
    const prev = clock.current;
    const result = clock.receiveTick(100);
    expect(result).toBe(prev + 1);
  });

  it('current returns counter', () => {
    const clock = new LamportClock('node-1');
    expect(clock.current).toBeGreaterThan(0);
  });

  it('nodeId is set correctly', () => {
    const clock = new LamportClock('my-node');
    expect(clock.nodeId).toBe('my-node');
  });
});

describe('mergeLWW', () => {
  it('remote wins with higher timestamp', () => {
    const local: LWWValue = { value: 'old', timestamp: 100, nodeId: 'a' };
    const remote: LWWValue = { value: 'new', timestamp: 200, nodeId: 'b' };
    expect(mergeLWW(local, remote)).toBe(true);
    expect(local.value).toBe('new');
    expect(local.timestamp).toBe(200);
  });

  it('remote loses with lower timestamp', () => {
    const local: LWWValue = { value: 'current', timestamp: 200, nodeId: 'a' };
    const remote: LWWValue = { value: 'old', timestamp: 100, nodeId: 'b' };
    expect(mergeLWW(local, remote)).toBe(false);
    expect(local.value).toBe('current');
  });

  it('same timestamp higher nodeId wins', () => {
    const local: LWWValue = { value: 'a-val', timestamp: 100, nodeId: 'aaa' };
    const remote: LWWValue = { value: 'b-val', timestamp: 100, nodeId: 'bbb' };
    expect(mergeLWW(local, remote)).toBe(true);
    expect(local.value).toBe('b-val');
  });

  it('same timestamp lower nodeId loses', () => {
    const local: LWWValue = { value: 'b-val', timestamp: 100, nodeId: 'bbb' };
    const remote: LWWValue = { value: 'a-val', timestamp: 100, nodeId: 'aaa' };
    expect(mergeLWW(local, remote)).toBe(false);
    expect(local.value).toBe('b-val');
  });
});

describe('DocumentCRDT', () => {
  const DOC_ID = 'doc-1';
  const NODE_A = 'node-a';
  const NODE_B = 'node-b';

  function makeUpdate(overrides: Partial<CRDTUpdateEvent> = {}): CRDTUpdateEvent {
    return {
      documentId: DOC_ID,
      blockId: 'block-1',
      field: 'content',
      value: 'hello',
      timestamp: 100,
      nodeId: NODE_A,
      ...overrides,
    };
  }

  it('applyRemoteUpdate creates block state', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    const accepted = crdt.applyRemoteUpdate(makeUpdate());
    expect(accepted).toBe(true);
    const state = crdt.getBlockState('block-1');
    expect(state).toBeDefined();
    expect(state?.content.value).toBe('hello');
  });

  it('later update wins', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    crdt.applyRemoteUpdate(makeUpdate({ value: 'first', timestamp: 100 }));
    const accepted = crdt.applyRemoteUpdate(makeUpdate({ value: 'second', timestamp: 200, nodeId: NODE_B }));
    expect(accepted).toBe(true);
    expect(crdt.getBlockState('block-1')?.content.value).toBe('second');
  });

  it('earlier update loses', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    crdt.applyRemoteUpdate(makeUpdate({ value: 'later', timestamp: 200 }));
    const accepted = crdt.applyRemoteUpdate(makeUpdate({ value: 'earlier', timestamp: 100, nodeId: NODE_B }));
    expect(accepted).toBe(false);
    expect(crdt.getBlockState('block-1')?.content.value).toBe('later');
  });

  it('update on deleted block is rejected', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    crdt.applyRemoteDelete(makeUpdate({ action: 'delete', timestamp: 100 }));
    const accepted = crdt.applyRemoteUpdate(makeUpdate({ timestamp: 200 }));
    expect(accepted).toBe(false);
  });

  it('checked field update works', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    const accepted = crdt.applyRemoteUpdate(makeUpdate({ field: 'checked', value: 'true' }));
    expect(accepted).toBe(true);
    expect(crdt.getBlockState('block-1')?.checked.value).toBe('true');
  });

  it('unknown field is rejected', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    const accepted = crdt.applyRemoteUpdate(makeUpdate({ field: 'unknown' }));
    expect(accepted).toBe(false);
  });

  it('applyRemoteDelete marks block as deleted', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    crdt.applyRemoteUpdate(makeUpdate());
    const accepted = crdt.applyRemoteDelete(makeUpdate({ action: 'delete', timestamp: 200 }));
    expect(accepted).toBe(true);
    expect(crdt.getBlockState('block-1')?.deleted).toBe(true);
  });

  it('duplicate delete is rejected', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    crdt.applyRemoteDelete(makeUpdate({ action: 'delete', timestamp: 200 }));
    const accepted = crdt.applyRemoteDelete(makeUpdate({ action: 'delete', timestamp: 100 }));
    expect(accepted).toBe(false);
  });

  it('delete on nonexistent block creates it deleted', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    const accepted = crdt.applyRemoteDelete(makeUpdate({ blockId: 'new-block', action: 'delete' }));
    expect(accepted).toBe(true);
    expect(crdt.getBlockState('new-block')?.deleted).toBe(true);
  });

  it('createUpdateEvent advances clock and tracks state', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    const event = crdt.createUpdateEvent('block-1', 'content', 'edited');
    expect(event.documentId).toBe(DOC_ID);
    expect(event.blockId).toBe('block-1');
    expect(event.field).toBe('content');
    expect(event.value).toBe('edited');
    expect(event.action).toBe('update');
    expect(event.nodeId).toBe(NODE_A);
    expect(event.timestamp).toBeGreaterThan(0);

    const state = crdt.getBlockState('block-1');
    expect(state?.content.value).toBe('edited');
  });

  it('createUpdateEvent for checked field', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    const event = crdt.createUpdateEvent('block-1', 'checked', 'true');
    expect(event.field).toBe('checked');
    const state = crdt.getBlockState('block-1');
    expect(state?.checked.value).toBe('true');
  });

  it('createUpdateEvent for other field does not crash', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    const event = crdt.createUpdateEvent('block-1', 'other', 'val');
    expect(event.field).toBe('other');
  });

  it('createDeleteEvent marks block deleted and returns event', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    const event = crdt.createDeleteEvent('block-1');
    expect(event.action).toBe('delete');
    expect(event.blockId).toBe('block-1');
    expect(crdt.getBlockState('block-1')?.deleted).toBe(true);
  });

  it('createDeleteEvent on nonexistent block creates state', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    crdt.createDeleteEvent('new-block');
    expect(crdt.getBlockState('new-block')?.deleted).toBe(true);
  });

  it('getBlockState returns undefined for unknown block', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    expect(crdt.getBlockState('nonexistent')).toBeUndefined();
  });

  it('clear removes all blocks', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    crdt.applyRemoteUpdate(makeUpdate());
    crdt.clear();
    expect(crdt.getBlockState('block-1')).toBeUndefined();
  });

  it('documentId is set correctly', () => {
    const crdt = new DocumentCRDT(DOC_ID, NODE_A);
    expect(crdt.documentId).toBe(DOC_ID);
  });
});
