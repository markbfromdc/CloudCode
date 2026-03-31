import { useState, useEffect, useCallback } from 'react';
import { GitBranch as GitBranchIcon, GitCommit as GitCommitIcon, Plus, RotateCw, Check, FileText, FilePlus, FileMinus, FileQuestion } from 'lucide-react';
import {
  getGitStatus,
  getGitLog,
  getGitBranches,
  stageFiles,
  createCommit,
  type GitFileStatus,
  type GitCommit,
  type GitBranch,
} from '../../services/git';

function statusIcon(status: string) {
  switch (status.toLowerCase()) {
    case 'modified':
    case 'M':
      return <FileText size={14} className="text-[var(--warning)]" />;
    case 'added':
    case 'A':
      return <FilePlus size={14} className="text-[var(--success)]" />;
    case 'deleted':
    case 'D':
      return <FileMinus size={14} className="text-[var(--error)]" />;
    default:
      return <FileQuestion size={14} className="text-[var(--text-secondary)]" />;
  }
}

function statusLabel(status: string): string {
  const s = status.toLowerCase();
  if (s === 'm' || s === 'modified') return 'M';
  if (s === 'a' || s === 'added') return 'A';
  if (s === 'd' || s === 'deleted') return 'D';
  if (s === '?' || s === '??' || s === 'untracked') return 'U';
  return status.charAt(0).toUpperCase();
}

export default function GitPanel() {
  const [commitMessage, setCommitMessage] = useState('');
  const [changedFiles, setChangedFiles] = useState<GitFileStatus[]>([]);
  const [branches, setBranches] = useState<GitBranch[]>([]);
  const [currentBranch, setCurrentBranch] = useState('main');
  const [commits, setCommits] = useState<GitCommit[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const refresh = useCallback(async () => {
    setIsLoading(true);
    try {
      const [statusData, branchData, logData] = await Promise.all([
        getGitStatus(),
        getGitBranches(),
        getGitLog(),
      ]);
      setChangedFiles(statusData);
      setBranches(branchData);
      const current = branchData.find((b) => b.current);
      if (current) {
        setCurrentBranch(current.name);
      }
      setCommits(logData);
    } catch {
      // API not available - keep defaults
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const handleStageAll = async () => {
    try {
      await stageFiles([]);
      await refresh();
    } catch {
      // Stage failed
    }
  };

  const handleCommit = async () => {
    if (!commitMessage.trim()) return;
    try {
      await createCommit(commitMessage.trim());
      setCommitMessage('');
      await refresh();
    } catch {
      // Commit failed
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-2 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold">
        Source Control
        <button
          onClick={refresh}
          disabled={isLoading}
          className="hover:text-[var(--text-primary)] disabled:opacity-50"
          title="Refresh"
        >
          <RotateCw size={14} className={isLoading ? 'animate-spin' : ''} />
        </button>
      </div>
      <div className="px-3 space-y-3">
        <div className="flex items-center gap-2 bg-[var(--bg-tertiary)] border border-[var(--border)] rounded px-2 py-1.5">
          <input
            type="text"
            placeholder="Message (Ctrl+Enter to commit)"
            value={commitMessage}
            onChange={(e) => setCommitMessage(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && e.ctrlKey) {
                handleCommit();
              }
            }}
            className="flex-1 bg-transparent text-sm text-[var(--text-primary)] outline-none"
          />
        </div>
        <div className="flex gap-1">
          <button
            onClick={handleCommit}
            disabled={!commitMessage.trim()}
            className="flex-1 flex items-center justify-center gap-1.5 bg-[var(--accent)] hover:bg-[var(--accent-hover)] disabled:opacity-50 text-white text-xs py-1.5 rounded transition-colors"
          >
            <Check size={14} />
            Commit
          </button>
        </div>
        <div className="flex items-center gap-2 text-xs text-[var(--text-secondary)] px-1">
          <GitBranchIcon size={14} />
          <span>{currentBranch}</span>
        </div>
        <div className="border-t border-[var(--border)] pt-2">
          <div className="flex items-center justify-between px-1 py-1 text-xs text-[var(--text-secondary)]">
            <span className="uppercase tracking-wider font-semibold">
              Changes {changedFiles.length > 0 && `(${changedFiles.length})`}
            </span>
            <div className="flex gap-1">
              <button onClick={handleStageAll} title="Stage All" className="hover:text-[var(--text-primary)]">
                <Plus size={14} />
              </button>
              <button onClick={refresh} title="Refresh" className="hover:text-[var(--text-primary)]">
                <RotateCw size={14} />
              </button>
            </div>
          </div>
          {changedFiles.length === 0 ? (
            <div className="text-xs text-[var(--text-secondary)] px-1 py-4 text-center">
              No changes detected.
            </div>
          ) : (
            <div className="space-y-0.5 mt-1">
              {changedFiles.map((file) => (
                <div
                  key={file.path}
                  className="flex items-center gap-2 px-1 py-0.5 text-xs text-[var(--text-primary)] hover:bg-[var(--bg-hover)] rounded"
                >
                  {statusIcon(file.status)}
                  <span className="flex-1 truncate">{file.path}</span>
                  <span className="text-[var(--text-secondary)] font-mono">
                    {statusLabel(file.status)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
        <div className="border-t border-[var(--border)] pt-2">
          <div className="px-1 py-1 text-xs text-[var(--text-secondary)] uppercase tracking-wider font-semibold">
            Commits
          </div>
          {commits.length === 0 ? (
            <div className="text-xs text-[var(--text-secondary)] px-1 py-4 text-center">
              No commits yet.
            </div>
          ) : (
            <div className="space-y-1 mt-1">
              {commits.slice(0, 20).map((commit) => (
                <div
                  key={commit.hash}
                  className="flex items-center gap-2 px-1 py-1 text-xs text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] rounded"
                >
                  <GitCommitIcon size={14} className="shrink-0" />
                  <div className="flex-1 min-w-0">
                    <div className="truncate text-[var(--text-primary)]">{commit.message}</div>
                    <div className="truncate text-[10px]">
                      {commit.hash.slice(0, 7)} - {commit.author}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
