import { Terminal, FileOutput, AlertTriangle, ChevronDown, X, Maximize2, Minimize2 } from 'lucide-react';
import { useState } from 'react';
import { useWorkspace } from '../../context/WorkspaceContext';
import type { PanelTab } from '../../types';
import TerminalPanel from '../terminal/TerminalPanel';

const tabs: { id: PanelTab; label: string; icon: typeof Terminal }[] = [
  { id: 'terminal', label: 'Terminal', icon: Terminal },
  { id: 'output', label: 'Output', icon: FileOutput },
  { id: 'problems', label: 'Problems', icon: AlertTriangle },
];

export default function BottomPanel() {
  const { state, dispatch } = useWorkspace();
  const [isMaximized, setIsMaximized] = useState(false);

  if (!state.isPanelOpen) return null;

  return (
    <div
      className={`flex flex-col border-t border-[var(--border)] bg-[var(--bg-primary)]
        ${isMaximized ? 'flex-1' : 'h-64'}`}
    >
      <div className="flex items-center justify-between px-2 bg-[var(--bg-secondary)] border-b border-[var(--border)] shrink-0">
        <div className="flex items-center">
          {tabs.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              onClick={() => dispatch({ type: 'SET_PANEL', panel: id })}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-xs uppercase tracking-wider transition-colors
                ${state.activePanel === id
                  ? 'text-[var(--text-active)] border-b-2 border-b-[var(--accent)]'
                  : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)]'
                }`}
            >
              <Icon size={14} />
              {label}
            </button>
          ))}
        </div>
        <div className="flex items-center gap-0.5">
          <button
            onClick={() => setIsMaximized(!isMaximized)}
            className="p-1 text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
            title={isMaximized ? 'Restore' : 'Maximize'}
          >
            {isMaximized ? <Minimize2 size={14} /> : <Maximize2 size={14} />}
          </button>
          <button
            onClick={() => dispatch({ type: 'TOGGLE_PANEL' })}
            className="p-1 text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
            title="Close Panel"
          >
            <ChevronDown size={14} />
          </button>
        </div>
      </div>
      <div className="flex-1 overflow-hidden">
        {state.activePanel === 'terminal' && <TerminalPanel />}
        {state.activePanel === 'output' && (
          <div className="p-3 text-xs text-[var(--text-secondary)] font-mono">
            <p>[CloudCode] Ready.</p>
          </div>
        )}
        {state.activePanel === 'problems' && (
          <div className="p-3 text-xs text-[var(--text-secondary)]">
            No problems detected.
          </div>
        )}
      </div>
    </div>
  );
}
