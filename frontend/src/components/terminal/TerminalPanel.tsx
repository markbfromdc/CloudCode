import { useEffect, useRef, useCallback } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import '@xterm/xterm/css/xterm.css';
import { TerminalWebSocket } from '../../services/websocket';
import { useWorkspace } from '../../context/WorkspaceContext';

export default function TerminalPanel() {
  const { state, dispatch } = useWorkspace();
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<Terminal | null>(null);
  const wsRef = useRef<TerminalWebSocket | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);

  const handleMessage = useCallback((data: ArrayBuffer | string) => {
    if (!termRef.current) return;
    if (data instanceof ArrayBuffer) {
      termRef.current.write(new Uint8Array(data));
    } else {
      termRef.current.write(data);
    }
  }, []);

  const handleStatus = useCallback((connected: boolean) => {
    dispatch({ type: 'SET_CONNECTED', connected });
  }, [dispatch]);

  useEffect(() => {
    if (!containerRef.current) return;

    const term = new Terminal({
      theme: {
        background: '#1e1e1e',
        foreground: '#cccccc',
        cursor: '#cccccc',
        cursorAccent: '#1e1e1e',
        selectionBackground: '#264f78',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#e5e5e5',
      },
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, monospace",
      fontSize: 13,
      lineHeight: 1.4,
      cursorBlink: true,
      cursorStyle: 'bar',
      allowProposedApi: true,
      scrollback: 10000,
    });

    const fitAddon = new FitAddon();
    const webLinksAddon = new WebLinksAddon();

    term.loadAddon(fitAddon);
    term.loadAddon(webLinksAddon);

    term.open(containerRef.current);
    fitAddon.fit();

    termRef.current = term;
    fitAddonRef.current = fitAddon;

    // Connect WebSocket if we have a session.
    if (state.sessionId) {
      const ws = new TerminalWebSocket(state.sessionId, handleMessage, handleStatus);
      wsRef.current = ws;
      ws.connect();

      term.onData((data: string) => ws.send(data));
      term.onResize(({ cols, rows }) => ws.resize(cols, rows));
    } else {
      // Demo mode: local echo terminal.
      term.writeln('\x1b[1;36mCloudCode IDE Terminal\x1b[0m');
      term.writeln('\x1b[90mCreate a workspace to connect to a live container.\x1b[0m');
      term.writeln('');
      term.write('\x1b[1;32m$\x1b[0m ');

      term.onData((data: string) => {
        if (data === '\r') {
          term.writeln('');
          term.write('\x1b[1;32m$\x1b[0m ');
        } else if (data === '\x7f') {
          // Backspace
          term.write('\b \b');
        } else {
          term.write(data);
        }
      });
    }

    const resizeObserver = new ResizeObserver(() => {
      try {
        fitAddon.fit();
      } catch {
        // Ignore resize errors during teardown.
      }
    });
    resizeObserver.observe(containerRef.current);

    return () => {
      resizeObserver.disconnect();
      wsRef.current?.disconnect();
      term.dispose();
    };
  }, [state.sessionId, handleMessage, handleStatus]);

  return (
    <div
      ref={containerRef}
      className="h-full w-full bg-[var(--bg-primary)]"
    />
  );
}
