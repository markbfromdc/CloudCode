import { useWorkspace } from '../../context/WorkspaceContext';
import FileExplorer from './FileExplorer';
import SearchPanel from './SearchPanel';
import GitPanel from './GitPanel';

export default function Sidebar() {
  const { state } = useWorkspace();

  if (!state.isSidebarOpen) return null;

  return (
    <div className="w-64 bg-[var(--bg-sidebar)] border-r border-[var(--border)] flex flex-col shrink-0 overflow-hidden">
      {state.activeActivity === 'explorer' && <FileExplorer />}
      {state.activeActivity === 'search' && <SearchPanel />}
      {state.activeActivity === 'git' && <GitPanel />}
      {state.activeActivity === 'extensions' && (
        <div className="p-4 text-xs text-[var(--text-secondary)]">
          <div className="uppercase tracking-wider font-semibold mb-2">Extensions</div>
          <p>Marketplace coming soon.</p>
        </div>
      )}
      {state.activeActivity === 'settings' && (
        <div className="p-4 text-xs text-[var(--text-secondary)]">
          <div className="uppercase tracking-wider font-semibold mb-2">Settings</div>
          <p>Settings editor coming soon.</p>
        </div>
      )}
    </div>
  );
}
