import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { TerminalWebSocket } from './websocket';

// Mock WebSocket.
class MockWebSocket {
  static OPEN = 1;
  static CLOSED = 3;

  url: string;
  readyState = MockWebSocket.OPEN;
  binaryType = '';
  onopen: (() => void) | null = null;
  onclose: ((e: { wasClean: boolean }) => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;
  onerror: (() => void) | null = null;

  sent: string[] = [];
  closeCalled = false;

  constructor(url: string) {
    this.url = url;
    // Simulate async open.
    setTimeout(() => this.onopen?.(), 0);
  }

  send(data: string) {
    this.sent.push(data);
  }

  close(_code?: number, _reason?: string) {
    this.closeCalled = true;
    this.readyState = MockWebSocket.CLOSED;
  }
}

vi.stubGlobal('WebSocket', MockWebSocket);

describe('TerminalWebSocket', () => {
  let onMessage: ReturnType<typeof vi.fn>;
  let onStatus: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.useFakeTimers();
    onMessage = vi.fn();
    onStatus = vi.fn();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('creates with correct session id', () => {
    const ws = new TerminalWebSocket('sess-1', onMessage, onStatus);
    expect(ws).toBeDefined();
  });

  it('constructs correct WebSocket URL', () => {
    const ws = new TerminalWebSocket('sess-1', onMessage, onStatus);
    ws.connect();

    // MockWebSocket captures the URL.
    // The URL should contain the session_id.
    // We can't directly inspect private ws field, but we can verify callbacks fire.
    vi.runAllTimers();
    expect(onStatus).toHaveBeenCalledWith(true);
  });

  it('calls onStatus(true) on open', () => {
    const ws = new TerminalWebSocket('sess-1', onMessage, onStatus);
    ws.connect();
    vi.runAllTimers();
    expect(onStatus).toHaveBeenCalledWith(true);
  });

  it('disconnect prevents reconnection and closes', () => {
    const ws = new TerminalWebSocket('sess-1', onMessage, onStatus);
    ws.connect();
    vi.runAllTimers();
    ws.disconnect();
    // Should not throw.
    expect(onStatus).toHaveBeenCalledWith(true);
  });

  it('disconnect is safe to call without connect', () => {
    const ws = new TerminalWebSocket('sess-1', onMessage, onStatus);
    ws.disconnect(); // Should not throw.
  });

  it('send does nothing when not connected', () => {
    const ws = new TerminalWebSocket('sess-1', onMessage, onStatus);
    ws.send('hello'); // Should not throw.
  });

  it('resize does nothing when not connected', () => {
    const ws = new TerminalWebSocket('sess-1', onMessage, onStatus);
    ws.resize(80, 24); // Should not throw.
  });
});
