import { useEffect } from 'react';
import { useWorkspace } from '../context/WorkspaceContext';

/** Registers global keyboard shortcuts for the IDE. */
export function useKeyboardShortcuts(onCommandPalette: () => void) {
  const { state, dispatch } = useWorkspace();

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const ctrl = e.ctrlKey || e.metaKey;
      const shift = e.shiftKey;

      // Ctrl+P: Quick Open (Command Palette in file mode)
      if (ctrl && !shift && e.key === 'p') {
        e.preventDefault();
        onCommandPalette();
        return;
      }

      // Ctrl+Shift+P: Command Palette (command mode)
      if (ctrl && shift && e.key === 'P') {
        e.preventDefault();
        onCommandPalette();
        return;
      }

      // Ctrl+B: Toggle Sidebar
      if (ctrl && !shift && e.key === 'b') {
        e.preventDefault();
        dispatch({ type: 'TOGGLE_SIDEBAR' });
        return;
      }

      // Ctrl+`: Toggle Terminal
      if (ctrl && e.key === '`') {
        e.preventDefault();
        dispatch({ type: 'TOGGLE_PANEL' });
        return;
      }

      // Ctrl+W: Close active tab
      if (ctrl && !shift && e.key === 'w') {
        e.preventDefault();
        if (state.activeTabId) {
          dispatch({ type: 'CLOSE_TAB', tabId: state.activeTabId });
        }
        return;
      }

      // Ctrl+Shift+F: Open Search
      if (ctrl && shift && e.key === 'F') {
        e.preventDefault();
        dispatch({ type: 'SET_ACTIVITY', activity: 'search' });
        return;
      }

      // Ctrl+Shift+G: Open Git
      if (ctrl && shift && e.key === 'G') {
        e.preventDefault();
        dispatch({ type: 'SET_ACTIVITY', activity: 'git' });
        return;
      }

      // Ctrl+Shift+E: Open Explorer
      if (ctrl && shift && e.key === 'E') {
        e.preventDefault();
        dispatch({ type: 'SET_ACTIVITY', activity: 'explorer' });
        return;
      }

      // Ctrl+1-9: Switch to tab by index
      if (ctrl && !shift && e.key >= '1' && e.key <= '9') {
        e.preventDefault();
        const idx = parseInt(e.key) - 1;
        if (idx < state.openTabs.length) {
          dispatch({ type: 'SET_ACTIVE_TAB', tabId: state.openTabs[idx].id });
        }
        return;
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [dispatch, state.activeTabId, state.openTabs, onCommandPalette]);
}
