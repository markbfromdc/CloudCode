import { ChevronRight, ChevronDown, File, Folder, FolderOpen } from 'lucide-react';
import { useWorkspace } from '../../context/WorkspaceContext';
import { getLanguageForFile } from '../../hooks/useFileLanguage';
import type { FileNode, EditorTab } from '../../types';

/** Returns a color class based on file extension for visual distinction. */
function getFileColor(name: string): string {
  const ext = name.split('.').pop()?.toLowerCase() || '';
  const colors: Record<string, string> = {
    ts: 'text-blue-400',
    tsx: 'text-blue-400',
    js: 'text-yellow-400',
    jsx: 'text-yellow-400',
    json: 'text-yellow-300',
    css: 'text-purple-400',
    html: 'text-orange-400',
    py: 'text-green-400',
    go: 'text-cyan-400',
    md: 'text-gray-400',
  };
  return colors[ext] || 'text-[var(--text-secondary)]';
}

function FileTreeItem({ node, depth }: { node: FileNode; depth: number }) {
  const { state, dispatch } = useWorkspace();

  const handleClick = () => {
    if (node.type === 'directory') {
      dispatch({ type: 'TOGGLE_FILE_EXPAND', path: node.path });
    } else {
      const tab: EditorTab = {
        id: node.path,
        path: node.path,
        name: node.name,
        language: getLanguageForFile(node.name),
        content: '',
        isDirty: false,
      };
      dispatch({ type: 'OPEN_FILE', tab });
    }
  };

  const isActive = state.activeTabId === node.path;

  return (
    <>
      <button
        onClick={handleClick}
        className={`flex items-center w-full text-left text-sm py-px hover:bg-[var(--bg-hover)] cursor-pointer
          ${isActive ? 'bg-[var(--bg-active)]' : ''}`}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
      >
        {node.type === 'directory' ? (
          <>
            {node.isExpanded ? (
              <ChevronDown size={16} className="shrink-0 text-[var(--text-secondary)]" />
            ) : (
              <ChevronRight size={16} className="shrink-0 text-[var(--text-secondary)]" />
            )}
            {node.isExpanded ? (
              <FolderOpen size={16} className="shrink-0 ml-0.5 mr-1.5 text-[var(--warning)]" />
            ) : (
              <Folder size={16} className="shrink-0 ml-0.5 mr-1.5 text-[var(--warning)]" />
            )}
          </>
        ) : (
          <>
            <span className="w-4 shrink-0" />
            <File size={16} className={`shrink-0 ml-0.5 mr-1.5 ${getFileColor(node.name)}`} />
          </>
        )}
        <span className="truncate text-[var(--text-primary)]">{node.name}</span>
      </button>
      {node.type === 'directory' && node.isExpanded && node.children?.map((child) => (
        <FileTreeItem key={child.path} node={child} depth={depth + 1} />
      ))}
    </>
  );
}

export default function FileExplorer() {
  const { state } = useWorkspace();

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-2 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold">
        Explorer
      </div>
      <div className="flex items-center px-4 py-1.5 text-xs uppercase tracking-wider text-[var(--text-primary)] font-semibold bg-[var(--bg-sidebar)]">
        Workspace
      </div>
      <div className="flex-1 overflow-y-auto">
        {state.files.length === 0 ? (
          <div className="px-4 py-8 text-xs text-[var(--text-secondary)] text-center">
            No files in workspace.<br />Create a workspace to get started.
          </div>
        ) : (
          state.files.map((node) => (
            <FileTreeItem key={node.path} node={node} depth={0} />
          ))
        )}
      </div>
    </div>
  );
}
