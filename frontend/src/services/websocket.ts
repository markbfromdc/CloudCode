/** WebSocket service for terminal communication with the backend. */

type MessageHandler = (data: ArrayBuffer | string) => void;
type StatusHandler = (connected: boolean) => void;

export class TerminalWebSocket {
  private ws: WebSocket | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private baseReconnectDelay = 1000;
  private onMessage: MessageHandler;
  private onStatus: StatusHandler;
  private sessionId: string;

  constructor(sessionId: string, onMessage: MessageHandler, onStatus: StatusHandler) {
    this.sessionId = sessionId;
    this.onMessage = onMessage;
    this.onStatus = onStatus;
  }

  /** Establish the WebSocket connection. */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${protocol}//${window.location.host}/ws/terminal?session_id=${this.sessionId}`;

    this.ws = new WebSocket(url);
    this.ws.binaryType = 'arraybuffer';

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.onStatus(true);
    };

    this.ws.onmessage = (event: MessageEvent) => {
      this.onMessage(event.data);
    };

    this.ws.onclose = (event: CloseEvent) => {
      this.onStatus(false);
      if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = () => {
      this.onStatus(false);
    };
  }

  /** Send data (keystrokes) to the terminal. */
  send(data: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(data);
    }
  }

  /** Send a terminal resize event. */
  resize(cols: number, rows: number): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'resize', cols, rows }));
    }
  }

  /** Close the WebSocket connection. */
  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection.
    if (this.ws) {
      this.ws.close(1000, 'client disconnect');
      this.ws = null;
    }
  }

  private scheduleReconnect(): void {
    const delay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectAttempts++;
    this.reconnectTimer = setTimeout(() => this.connect(), delay);
  }
}
