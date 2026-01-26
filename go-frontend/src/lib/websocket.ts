import { Message } from './api';

type EventType = 'message' | 'typing' | 'read' | 'online' | 'offline' | 'conversation' | 'message_edit' | 'message_delete';

interface WSEvent {
  type: EventType;
  conversation_id?: number;
  user_id?: number;
  payload?: unknown;
  timestamp?: string;
}

type EventHandler = (event: WSEvent, data?: Message) => void;

class WebSocketClient {
  private ws: WebSocket | null = null;
  private handlers: Map<EventType, Set<EventHandler>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isConnecting = false;

  connect() {
    if (this.isConnecting || this.ws?.readyState === WebSocket.OPEN) return;
    
    this.isConnecting = true;
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';
    
    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.isConnecting = false;
      };

      this.ws.onmessage = (event) => {
        try {
          const wsEvent: WSEvent = JSON.parse(event.data);
          const handlers = this.handlers.get(wsEvent.type);
          if (handlers) {
            // Payload is already parsed as part of the JSON
            const data = wsEvent.payload as Message | undefined;
            handlers.forEach((handler) => handler(wsEvent, data));
          }
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e, event.data);
        }
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.isConnecting = false;
        this.tryReconnect();
      };

      this.ws.onerror = () => {
        this.isConnecting = false;
      };
    } catch {
      this.isConnecting = false;
    }
  }

  private tryReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => this.connect(), this.reconnectDelay * this.reconnectAttempts);
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  on(event: EventType, handler: EventHandler) {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, new Set());
    }
    this.handlers.get(event)!.add(handler);
    return () => this.handlers.get(event)?.delete(handler);
  }

  send(data: object) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }
}

export const wsClient = new WebSocketClient();
