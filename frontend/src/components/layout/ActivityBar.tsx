import { Files, Search, GitBranch, Blocks, Settings } from 'lucide-react';
import { useWorkspace } from '../../context/WorkspaceContext';
import type { ActivityBarItem } from '../../types';

const activities: { id: ActivityBarItem; icon: typeof Files; label: string }[] = [
  { id: 'explorer', icon: Files, label: 'Explorer' },
  { id: 'search', icon: Search, label: 'Search' },
  { id: 'git', icon: GitBranch, label: 'Source Control' },
  { id: 'extensions', icon: Blocks, label: 'Extensions' },
];

export default function ActivityBar() {
  const { state, dispatch } = useWorkspace();

  return (
    <div className="flex flex-col items-center w-12 bg-[var(--bg-sidebar)] border-r border-[var(--border)] shrink-0">
      {activities.map(({ id, icon: Icon, label }) => (
        <button
          key={id}
          title={label}
          onClick={() => dispatch({ type: 'SET_ACTIVITY', activity: id })}
          className={`w-12 h-12 flex items-center justify-center transition-colors relative
            ${state.activeActivity === id && state.isSidebarOpen
              ? 'text-[var(--text-active)] before:absolute before:left-0 before:top-1/2 before:-translate-y-1/2 before:w-0.5 before:h-6 before:bg-[var(--text-active)]'
              : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)]'
            }`}
        >
          <Icon size={22} />
        </button>
      ))}
      <div className="flex-1" />
      <button
        title="Settings"
        onClick={() => dispatch({ type: 'SET_ACTIVITY', activity: 'settings' })}
        className="w-12 h-12 flex items-center justify-center text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
      >
        <Settings size={22} />
      </button>
    </div>
  );
}
