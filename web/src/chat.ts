/**
 * Sovereign Chat - Typed WebSocket Interface
 */

export interface DerivationStarted {
  type: 'DERIVATION_STARTED';
  derivationId: string;
}

export interface DerivationStep {
  type: 'DERIVATION_STEP';
  derivationId: string;
  step: string;
  payload: string;
}

export interface DerivationCompleted {
  type: 'DERIVATION_COMPLETED';
  derivationId: string;
  result: string;
}

export interface AuditTrailEntry {
  type: 'AUDIT_TRAIL';
  entryId: string;
  timestamp: string;
  signature: string;
}

export type SovereignMessage = 
  | DerivationStarted 
  | DerivationStep 
  | DerivationCompleted 
  | AuditTrailEntry;

type MessageCallback<T> = (msg: T) => void;

export class TypedWebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private callbacks: Map<string, MessageCallback<any>[]> = new Map();
  private reconnectDelay = 1000;
  private maxReconnectDelay = 30000;

  constructor(url: string) {
    this.url = url;
  }

  public connect(): void {
    console.log(`[WS] Connecting to ${this.url}...`);
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log('[WS] Connected');
      this.reconnectDelay = 1000; // Reset delay
      this.emit('status', 'connected');
    };

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const msg: SovereignMessage = JSON.parse(event.data);
        this.emit(msg.type, msg);
      } catch (err) {
        console.error('[WS] Failed to parse message:', err);
      }
    };

    this.ws.onclose = () => {
      console.warn(`[WS] Connection closed. Reconnecting in ${this.reconnectDelay}ms...`);
      this.emit('status', 'reconnecting');
      setTimeout(() => {
        this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
        this.connect();
      }, this.reconnectDelay);
    };
  }

  public on<K extends SovereignMessage['type'] | 'status'>(
    type: K, 
    callback: MessageCallback<K extends 'status' ? string : Extract<SovereignMessage, { type: K }>>
  ): void {
    if (!this.callbacks.has(type)) {
      this.callbacks.set(type, []);
    }
    this.callbacks.get(type)!.push(callback);
  }

  private emit(type: string, data: any): void {
    const list = this.callbacks.get(type);
    if (list) {
      list.forEach(cb => cb(data));
    }
  }

  public send(msg: object): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  public disconnect(): void {
    if (this.ws) {
      this.ws.close();
    }
  }
}
