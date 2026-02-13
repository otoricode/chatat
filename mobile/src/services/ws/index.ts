// WebSocket client singleton
import WebSocketClient from './WebSocketClient';

export const wsClient = new WebSocketClient();
export { WebSocketClient };
export type { WSMessage, ConnectionState } from './WebSocketClient';
