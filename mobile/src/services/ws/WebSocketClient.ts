// WebSocket client with auto-reconnect, offline queue, and event system
import { AppState, type AppStateStatus } from 'react-native';

export type WSMessage = {
  type: string;
  payload: unknown;
};

type MessageHandler = (payload: unknown) => void;

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting';

class WebSocketClient {
  private ws: WebSocket | null = null;
  private url = '';
  private token = '';
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private maxReconnectDelay = 30000;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private pingTimer: ReturnType<typeof setInterval> | null = null;
  private messageHandlers: Map<string, Set<MessageHandler>> = new Map();
  private stateHandlers: Set<(state: ConnectionState) => void> = new Set();
  private pendingMessages: WSMessage[] = [];
  private _state: ConnectionState = 'disconnected';
  private intentionalClose = false;

  get state(): ConnectionState {
    return this._state;
  }

  private setState(state: ConnectionState) {
    this._state = state;
    for (const handler of this.stateHandlers) {
      handler(state);
    }
  }

  connect(url: string, token: string): void {
    this.url = url;
    this.token = token;
    this.intentionalClose = false;
    this.doConnect();

    // Handle app state changes (background/foreground)
    AppState.addEventListener('change', this.handleAppState);
  }

  disconnect(): void {
    this.intentionalClose = true;
    this.cleanup();
    this.setState('disconnected');
  }

  send(type: string, payload: unknown): void {
    const msg: WSMessage = { type, payload };

    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    } else {
      // Queue for later
      this.pendingMessages.push(msg);
    }
  }

  on(type: string, handler: MessageHandler): () => void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set());
    }
    this.messageHandlers.get(type)?.add(handler);

    // Return unsubscribe function
    return () => {
      this.messageHandlers.get(type)?.delete(handler);
    };
  }

  off(type: string, handler: MessageHandler): void {
    this.messageHandlers.get(type)?.delete(handler);
  }

  onStateChange(handler: (state: ConnectionState) => void): () => void {
    this.stateHandlers.add(handler);
    return () => {
      this.stateHandlers.delete(handler);
    };
  }

  // --- Internal ---

  private doConnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.setState('connecting');

    const wsUrl = `${this.url}?token=${this.token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => this.handleOpen();
    this.ws.onmessage = (event) => this.handleMessage(event);
    this.ws.onclose = (event) => this.handleClose(event);
    this.ws.onerror = () => {
      // onclose will be called after onerror
    };
  }

  private handleOpen(): void {
    this.reconnectAttempts = 0;
    this.reconnectDelay = 1000;
    this.setState('connected');
    this.startPing();
    this.flushPendingMessages();
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const msg = JSON.parse(event.data as string) as WSMessage;
      const handlers = this.messageHandlers.get(msg.type);
      if (handlers) {
        for (const handler of handlers) {
          try {
            handler(msg.payload);
          } catch {
            // Swallow handler errors
          }
        }
      }
    } catch {
      // Invalid JSON, ignore
    }
  }

  private handleClose(_event: CloseEvent): void {
    this.stopPing();
    this.ws = null;

    if (!this.intentionalClose) {
      this.setState('reconnecting');
      this.scheduleReconnect();
    } else {
      this.setState('disconnected');
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.setState('disconnected');
      return;
    }

    const delay = Math.min(this.reconnectDelay, this.maxReconnectDelay);
    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++;
      this.reconnectDelay *= 2;
      this.doConnect();
    }, delay);
  }

  private flushPendingMessages(): void {
    const pending = [...this.pendingMessages];
    this.pendingMessages = [];

    for (const msg of pending) {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify(msg));
      } else {
        this.pendingMessages.push(msg);
        break;
      }
    }
  }

  private startPing(): void {
    this.stopPing();
    // Send a ping every 30 seconds to keep connection alive
    this.pingTimer = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send('ping', {});
      }
    }, 30000);
  }

  private stopPing(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer);
      this.pingTimer = null;
    }
  }

  private cleanup(): void {
    this.stopPing();

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.onopen = null;
      this.ws.onmessage = null;
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.close();
      this.ws = null;
    }

    this.pendingMessages = [];
    this.reconnectAttempts = 0;
    this.reconnectDelay = 1000;
  }

  private handleAppState = (nextState: AppStateStatus): void => {
    if (nextState === 'active' && this._state === 'disconnected' && !this.intentionalClose && this.url) {
      // App came back to foreground, reconnect
      this.doConnect();
    }
  };
}

export default WebSocketClient;
