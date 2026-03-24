import { GitBranch, GitCommit, Plus, RotateCw, Check } from 'lucide-react';

export default function GitPanel() {
  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-2 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold">
        Source Control
      </div>
      <div className="px-3 space-y-3">
        <div className="flex items-center gap-2 bg-[var(--bg-tertiary)] border border-[var(--border)] rounded px-2 py-1.5">
          <input
            type="text"
            placeholder="Message (Ctrl+Enter to commit)"
            className="flex-1 bg-transparent text-sm text-[var(--text-primary)] outline-none"
          />
        </div>
        <div className="flex gap-1">
          <button className="flex-1 flex items-center justify-center gap-1.5 bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white text-xs py-1.5 rounded transition-colors">
            <Check size={14} />
            Commit
          </button>
        </div>
        <div className="flex items-center gap-2 text-xs text-[var(--text-secondary)] px-1">
          <GitBranch size={14} />
          <span>main</span>
          <span className="text-[var(--success)]">0 ahead</span>
          <span>0 behind</span>
        </div>
        <div className="border-t border-[var(--border)] pt-2">
          <div className="flex items-center justify-between px-1 py-1 text-xs text-[var(--text-secondary)]">
            <span className="uppercase tracking-wider font-semibold">Changes</span>
            <div className="flex gap-1">
              <button title="Stage All" className="hover:text-[var(--text-primary)]">
                <Plus size={14} />
              </button>
              <button title="Refresh" className="hover:text-[var(--text-primary)]">
                <RotateCw size={14} />
              </button>
            </div>
          </div>
          <div className="text-xs text-[var(--text-secondary)] px-1 py-4 text-center">
            No changes detected.
          </div>
        </div>
        <div className="border-t border-[var(--border)] pt-2">
          <div className="px-1 py-1 text-xs text-[var(--text-secondary)] uppercase tracking-wider font-semibold">
            Commits
          </div>
          <div className="space-y-1 mt-1">
            <div className="flex items-center gap-2 px-1 py-1 text-xs text-[var(--text-secondary)]">
              <GitCommit size={14} />
              <span className="truncate">Initial commit</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
