import { X, Circle } from 'lucide-react';
import { useWorkspace } from '../../context/WorkspaceContext';

export default function EditorTabs() {
  const { state, dispatch } = useWorkspace();

  if (state.openTabs.length === 0) return null;

  return (
    <div className="flex items-center bg-[var(--bg-secondary)] border-b border-[var(--border)] overflow-x-auto shrink-0">
      {state.openTabs.map((tab) => (
        <div
          key={tab.id}
          onClick={() => dispatch({ type: 'SET_ACTIVE_TAB', tabId: tab.id })}
          className={`flex items-center gap-1.5 px-3 py-1.5 text-sm cursor-pointer border-r border-[var(--border)] select-none group min-w-0
            ${tab.id === state.activeTabId
              ? 'bg-[var(--bg-primary)] text-[var(--text-active)] border-t-2 border-t-[var(--accent)]'
              : 'bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] border-t-2 border-t-transparent'
            }`}
        >
          <span className="truncate max-w-[120px]">{tab.name}</span>
          {tab.isDirty && (
            <Circle size={8} className="fill-current text-[var(--text-secondary)] shrink-0" />
          )}
          <button
            onClick={(e) => {
              e.stopPropagation();
              dispatch({ type: 'CLOSE_TAB', tabId: tab.id });
            }}
            className="opacity-0 group-hover:opacity-100 hover:bg-[var(--bg-active)] rounded p-0.5 shrink-0 transition-opacity"
          >
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}
