import { GitBranch, Circle } from 'lucide-react';
import { useWorkspace } from '../../context/WorkspaceContext';

export default function StatusBar() {
  const { state } = useWorkspace();

  const activeTab = state.openTabs.find((t) => t.id === state.activeTabId);

  return (
    <div className="flex items-center h-6 bg-[var(--accent)] text-white text-xs px-2 shrink-0 select-none">
      <div className="flex items-center gap-3">
        <span className="flex items-center gap-1">
          <GitBranch size={12} />
          main
        </span>
        <span className="flex items-center gap-1">
          <Circle
            size={8}
            className={state.isConnected ? 'fill-green-300 text-green-300' : 'fill-red-400 text-red-400'}
          />
          {state.isConnected ? 'Connected' : 'Disconnected'}
        </span>
      </div>
      <div className="flex-1" />
      <div className="flex items-center gap-3">
        {activeTab && (
          <>
            <span>{activeTab.language}</span>
            <span>UTF-8</span>
          </>
        )}
        <span>CloudCode IDE</span>
      </div>
    </div>
  );
}
