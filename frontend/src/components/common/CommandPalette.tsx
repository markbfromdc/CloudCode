import { useState, useEffect, useRef, useMemo } from 'react';
import { Search } from 'lucide-react';
import { useWorkspace } from '../../context/WorkspaceContext';
import { getLanguageForFile } from '../../hooks/useFileLanguage';
import type { EditorTab, FileNode } from '../../types';

interface Command {
  id: string;
  label: string;
  category: string;
  shortcut?: string;
  action: () => void;
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}

/** Recursively flattens a file tree into a list of file paths. */
function flattenFiles(nodes: FileNode[]): FileNode[] {
  const result: FileNode[] = [];
  for (const node of nodes) {
    if (node.type === 'file') {
      result.push(node);
    }
    if (node.children) {
      result.push(...flattenFiles(node.children));
    }
  }
  return result;
}

export default function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
  const { state, dispatch } = useWorkspace();
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const isCommandMode = query.startsWith('>');

  const commands: Command[] = useMemo(() => [
    {
      id: 'toggle-terminal',
      label: 'Toggle Terminal',
      category: 'View',
      shortcut: 'Ctrl+`',
      action: () => dispatch({ type: 'TOGGLE_PANEL' }),
    },
    {
      id: 'toggle-sidebar',
      label: 'Toggle Sidebar',
      category: 'View',
      shortcut: 'Ctrl+B',
      action: () => dispatch({ type: 'TOGGLE_SIDEBAR' }),
    },
    {
      id: 'open-explorer',
      label: 'Show Explorer',
      category: 'View',
      action: () => dispatch({ type: 'SET_ACTIVITY', activity: 'explorer' }),
    },
    {
      id: 'open-search',
      label: 'Show Search',
      category: 'View',
      shortcut: 'Ctrl+Shift+F',
      action: () => dispatch({ type: 'SET_ACTIVITY', activity: 'search' }),
    },
    {
      id: 'open-git',
      label: 'Show Source Control',
      category: 'View',
      shortcut: 'Ctrl+Shift+G',
      action: () => dispatch({ type: 'SET_ACTIVITY', activity: 'git' }),
    },
    {
      id: 'close-tab',
      label: 'Close Active Editor',
      category: 'File',
      shortcut: 'Ctrl+W',
      action: () => {
        if (state.activeTabId) {
          dispatch({ type: 'CLOSE_TAB', tabId: state.activeTabId });
        }
      },
    },
    {
      id: 'open-terminal',
      label: 'Open Terminal',
      category: 'Terminal',
      action: () => {
        dispatch({ type: 'SET_PANEL', panel: 'terminal' });
      },
    },
    {
      id: 'open-problems',
      label: 'Show Problems',
      category: 'View',
      action: () => dispatch({ type: 'SET_PANEL', panel: 'problems' }),
    },
  ], [dispatch, state.activeTabId]);

  const allFiles = useMemo(() => flattenFiles(state.files), [state.files]);

  const filtered = useMemo(() => {
    if (isCommandMode) {
      const cmdQuery = query.slice(1).toLowerCase();
      return commands
        .filter((c) => c.label.toLowerCase().includes(cmdQuery))
        .map((c) => ({
          id: c.id,
          label: c.label,
          detail: `${c.category}${c.shortcut ? ` (${c.shortcut})` : ''}`,
          action: c.action,
        }));
    }

    const q = query.toLowerCase();
    return allFiles
      .filter((f) => f.name.toLowerCase().includes(q) || f.path.toLowerCase().includes(q))
      .slice(0, 20)
      .map((f) => ({
        id: f.path,
        label: f.name,
        detail: f.path,
        action: () => {
          const tab: EditorTab = {
            id: f.path,
            path: f.path,
            name: f.name,
            language: getLanguageForFile(f.name),
            content: '',
            isDirty: false,
          };
          dispatch({ type: 'OPEN_FILE', tab });
        },
      }));
  }, [query, isCommandMode, commands, allFiles, dispatch]);

  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [isOpen]);

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Enter' && filtered[selectedIndex]) {
      filtered[selectedIndex].action();
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex justify-center pt-[15%]" onClick={onClose}>
      <div
        className="w-[560px] max-h-[400px] bg-[var(--bg-secondary)] border border-[var(--border)] rounded-lg shadow-2xl overflow-hidden flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-2 px-3 py-2 border-b border-[var(--border)]">
          <Search size={16} className="text-[var(--text-secondary)] shrink-0" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={isCommandMode ? 'Type a command...' : 'Search files by name (type > for commands)'}
            className="flex-1 bg-transparent text-sm text-[var(--text-primary)] outline-none"
          />
        </div>
        <div className="flex-1 overflow-y-auto">
          {filtered.length === 0 ? (
            <div className="px-4 py-6 text-sm text-[var(--text-secondary)] text-center">
              No matching results.
            </div>
          ) : (
            filtered.map((item, i) => (
              <button
                key={item.id}
                onClick={() => {
                  item.action();
                  onClose();
                }}
                className={`w-full flex items-center justify-between px-3 py-1.5 text-sm text-left
                  ${i === selectedIndex
                    ? 'bg-[var(--accent)] text-white'
                    : 'text-[var(--text-primary)] hover:bg-[var(--bg-hover)]'
                  }`}
              >
                <span className="truncate">{item.label}</span>
                <span className={`text-xs truncate ml-4 ${i === selectedIndex ? 'text-white/70' : 'text-[var(--text-secondary)]'}`}>
                  {item.detail}
                </span>
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
