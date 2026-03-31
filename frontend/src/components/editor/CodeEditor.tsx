import { useEffect, useCallback, useState } from 'react';
import Editor from '@monaco-editor/react';
import { useWorkspace } from '../../context/WorkspaceContext';
import { writeFile } from '../../services/api';

export default function CodeEditor() {
  const { state, dispatch } = useWorkspace();
  const [showSaved, setShowSaved] = useState(false);

  const activeTab = state.openTabs.find((t) => t.id === state.activeTabId);

  const handleSave = useCallback(() => {
    if (!activeTab || !activeTab.isDirty) return;
    if (state.sessionId) {
      writeFile(state.sessionId, activeTab.path, activeTab.content)
        .then(() => {
          dispatch({ type: 'MARK_TAB_SAVED', tabId: activeTab.id });
          setShowSaved(true);
          setTimeout(() => setShowSaved(false), 1500);
        })
        .catch(() => {
          // Save failed silently - tab stays dirty
        });
    } else {
      // No session - just mark as saved locally
      dispatch({ type: 'MARK_TAB_SAVED', tabId: activeTab.id });
      setShowSaved(true);
      setTimeout(() => setShowSaved(false), 1500);
    }
  }, [activeTab, state.sessionId, dispatch]);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        handleSave();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [handleSave]);

  if (!activeTab) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[var(--bg-primary)]">
        <div className="text-center text-[var(--text-secondary)]">
          <div className="text-6xl mb-4 opacity-20 font-bold">CC</div>
          <p className="text-lg mb-2">CloudCode IDE</p>
          <p className="text-sm">Open a file from the explorer to start editing</p>
          <div className="mt-8 text-xs space-y-1">
            <p><kbd className="px-1.5 py-0.5 bg-[var(--bg-tertiary)] rounded border border-[var(--border)]">Ctrl+P</kbd> Quick Open</p>
            <p><kbd className="px-1.5 py-0.5 bg-[var(--bg-tertiary)] rounded border border-[var(--border)]">Ctrl+Shift+P</kbd> Command Palette</p>
            <p><kbd className="px-1.5 py-0.5 bg-[var(--bg-tertiary)] rounded border border-[var(--border)]">Ctrl+`</kbd> Toggle Terminal</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-hidden relative">
      {showSaved && (
        <div className="absolute top-2 right-4 z-10 bg-[var(--success)] text-white text-xs px-2 py-1 rounded shadow">
          Saved
        </div>
      )}
      <Editor
        height="100%"
        language={activeTab.language}
        value={activeTab.content}
        theme="vs-dark"
        onChange={(value) => {
          if (value !== undefined) {
            dispatch({
              type: 'UPDATE_TAB_CONTENT',
              tabId: activeTab.id,
              content: value,
            });
          }
        }}
        options={{
          fontSize: 14,
          fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, monospace",
          fontLigatures: true,
          minimap: { enabled: true, scale: 1 },
          scrollBeyondLastLine: false,
          smoothScrolling: true,
          cursorBlinking: 'smooth',
          cursorSmoothCaretAnimation: 'on',
          renderLineHighlight: 'all',
          bracketPairColorization: { enabled: true },
          autoClosingBrackets: 'always',
          autoClosingQuotes: 'always',
          formatOnPaste: true,
          suggestOnTriggerCharacters: true,
          tabSize: 2,
          wordWrap: 'off',
          lineNumbers: 'on',
          glyphMargin: true,
          folding: true,
          links: true,
          padding: { top: 8 },
        }}
      />
    </div>
  );
}
