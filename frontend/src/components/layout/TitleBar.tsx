import { Menu, Minus, Square, X, Play, RotateCw } from 'lucide-react';
import { useWorkspace } from '../../context/WorkspaceContext';

export default function TitleBar() {
  const { state } = useWorkspace();
  const activeTab = state.openTabs.find((t) => t.id === state.activeTabId);

  return (
    <div className="flex items-center h-9 bg-[var(--bg-secondary)] border-b border-[var(--border)] select-none shrink-0 draggable">
      <div className="flex items-center gap-2 px-3">
        <button className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]">
          <Menu size={16} />
        </button>
      </div>

      <div className="flex items-center gap-1 px-2">
        <button
          className="flex items-center gap-1 px-2 py-0.5 text-xs text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] rounded"
        >
          File
        </button>
        <button className="flex items-center gap-1 px-2 py-0.5 text-xs text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] rounded">
          Edit
        </button>
        <button className="flex items-center gap-1 px-2 py-0.5 text-xs text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] rounded">
          View
        </button>
        <button className="flex items-center gap-1 px-2 py-0.5 text-xs text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] rounded">
          Terminal
        </button>
        <button className="flex items-center gap-1 px-2 py-0.5 text-xs text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] rounded">
          Help
        </button>
      </div>

      <div className="flex-1 flex justify-center">
        <span className="text-xs text-[var(--text-secondary)]">
          {activeTab ? `${activeTab.name} - ` : ''}CloudCode IDE
        </span>
      </div>

      <div className="flex items-center gap-1 px-2">
        <button className="p-1.5 text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)] rounded" title="Run">
          <Play size={14} />
        </button>
        <button className="p-1.5 text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)] rounded" title="Restart">
          <RotateCw size={14} />
        </button>
      </div>

      <div className="flex items-center">
        <button className="px-3 py-2 text-[var(--text-secondary)] hover:bg-[var(--bg-hover)]">
          <Minus size={14} />
        </button>
        <button className="px-3 py-2 text-[var(--text-secondary)] hover:bg-[var(--bg-hover)]">
          <Square size={12} />
        </button>
        <button className="px-3 py-2 text-[var(--text-secondary)] hover:bg-red-600 hover:text-white">
          <X size={14} />
        </button>
      </div>
    </div>
  );
}
