import { useWorkspace } from '../../context/WorkspaceContext';
import FileExplorer from './FileExplorer';
import SearchPanel from './SearchPanel';
import GitPanel from './GitPanel';
import ExtensionsPanel from './ExtensionsPanel';
import SettingsPanel from './SettingsPanel';

export default function Sidebar() {
  const { state } = useWorkspace();

  if (!state.isSidebarOpen) return null;

  return (
    <div className="w-64 bg-[var(--bg-sidebar)] border-r border-[var(--border)] flex flex-col shrink-0 overflow-hidden">
      {state.activeActivity === 'explorer' && <FileExplorer />}
      {state.activeActivity === 'search' && <SearchPanel />}
      {state.activeActivity === 'git' && <GitPanel />}
      {state.activeActivity === 'extensions' && <ExtensionsPanel />}
      {state.activeActivity === 'settings' && <SettingsPanel />}
    </div>
  );
}
