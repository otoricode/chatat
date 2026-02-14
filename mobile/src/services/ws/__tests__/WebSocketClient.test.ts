// @ts-nocheck
// Mock WebSocket and AppState
const mockAddEventListener = jest.fn();
jest.mock('react-native', () => ({
  AppState: {
    addEventListener: mockAddEventListener,
  },
}));

// Mock WebSocket
class MockWebSocket {
  static OPEN = 1;
  static CLOSED = 3;
  
  readyState = MockWebSocket.OPEN;
  onopen: (() => void) | null = null;
  onmessage: ((evt: { data: string }) => void) | null = null;
  onclose: ((evt: { code: number }) => void) | null = null;
  onerror: (() => void) | null = null;
  send = jest.fn();
  close = jest.fn();
}

(global as any).WebSocket = MockWebSocket;

import WebSocketClient from '../WebSocketClient';

let client: WebSocketClient;

beforeEach(() => {
  jest.clearAllMocks();
  jest.useFakeTimers();
  client = new WebSocketClient();
});

afterEach(() => {
  jest.useRealTimers();
});

describe('WebSocketClient', () => {
  it('starts disconnected', () => {
    expect(client.state).toBe('disconnected');
  });

  it('connect sets state to connecting', () => {
    client.connect('ws://test', 'token1');
    expect(client.state).toBe('connecting');
  });

  it('handles open event', () => {
    const stateHandler = jest.fn();
    client.onStateChange(stateHandler);
    client.connect('ws://test', 'token1');
    
    // Simulate open
    const ws = (client as any).ws;
    ws.onopen();
    
    expect(client.state).toBe('connected');
  });

  it('handles message events', () => {
    const handler = jest.fn();
    client.on('chat:message', handler);
    client.connect('ws://test', 'token1');
    
    const ws = (client as any).ws;
    ws.onopen();
    
    ws.onmessage({ data: JSON.stringify({ type: 'chat:message', payload: { text: 'hi' } }) });
    
    expect(handler).toHaveBeenCalledWith({ text: 'hi' });
  });

  it('ignores invalid JSON messages', () => {
    const handler = jest.fn();
    client.on('chat:message', handler);
    client.connect('ws://test', 'token1');
    
    const ws = (client as any).ws;
    ws.onopen();
    ws.onmessage({ data: 'not json' });
    
    expect(handler).not.toHaveBeenCalled();
  });

  it('send queues messages when not connected', () => {
    client.send('test', { data: 1 });
    expect((client as any).pendingMessages).toHaveLength(1);
  });

  it('send transmits when connected', () => {
    client.connect('ws://test', 'token1');
    const ws = (client as any).ws;
    ws.onopen();
    
    client.send('test', { data: 1 });
    expect(ws.send).toHaveBeenCalledWith(JSON.stringify({ type: 'test', payload: { data: 1 } }));
  });

  it('flushes pending messages on connection', () => {
    client.send('queued1', 'p1');
    client.send('queued2', 'p2');
    
    client.connect('ws://test', 'token1');
    const ws = (client as any).ws;
    ws.onopen();
    
    expect(ws.send).toHaveBeenCalledTimes(2);
  });

  it('on returns unsubscribe function', () => {
    const handler = jest.fn();
    const unsub = client.on('test', handler);
    
    unsub();
    
    client.connect('ws://test', 'token1');
    const ws = (client as any).ws;
    ws.onopen();
    ws.onmessage({ data: JSON.stringify({ type: 'test', payload: {} }) });
    
    expect(handler).not.toHaveBeenCalled();
  });

  it('off removes handler', () => {
    const handler = jest.fn();
    client.on('test', handler);
    client.off('test', handler);
    
    // Off for nonexistent type should not throw
    client.off('nonexistent', handler);
  });

  it('onStateChange returns unsubscribe', () => {
    const handler = jest.fn();
    const unsub = client.onStateChange(handler);
    
    client.connect('ws://test', 'token1');
    expect(handler).toHaveBeenCalledWith('connecting');
    
    unsub();
    handler.mockClear();
    
    const ws = (client as any).ws;
    ws.onopen();
    
    // handler should NOT be called after unsubscribe
    expect(handler).not.toHaveBeenCalled();
  });

  it('disconnect cleans up', () => {
    client.connect('ws://test', 'token1');
    const ws = (client as any).ws;
    
    client.disconnect();
    
    expect(ws.close).toHaveBeenCalled();
    expect(client.state).toBe('disconnected');
  });

  it('schedules reconnect on unintentional close', () => {
    client.connect('ws://test', 'token1');
    const ws = (client as any).ws;
    ws.onopen();
    
    ws.onclose({ code: 1006 });
    
    expect(client.state).toBe('reconnecting');
  });

  it('does not reconnect on intentional close', () => {
    client.connect('ws://test', 'token1');
    const ws = (client as any).ws;
    ws.onopen();
    
    client.disconnect();
    expect(client.state).toBe('disconnected');
  });

  it('stops reconnecting after max attempts', () => {
    client.connect('ws://test', 'token1');
    const ws = (client as any).ws;
    
    // Set reconnect attempts to max
    (client as any).reconnectAttempts = 10;
    
    ws.onclose({ code: 1006 });
    
    expect(client.state).toBe('disconnected');
  });

  it('handles handler errors gracefully', () => {
    const handler = jest.fn(() => { throw new Error('handler crash'); });
    client.on('test', handler);
    client.connect('ws://test', 'token1');
    
    const ws = (client as any).ws;
    ws.onopen();
    
    // Should not throw
    ws.onmessage({ data: JSON.stringify({ type: 'test', payload: {} }) });
    expect(handler).toHaveBeenCalled();
  });
});
