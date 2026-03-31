import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { WorkspaceProvider, useWorkspace } from './WorkspaceContext';
import type { EditorTab, FileNode } from '../types';
import type { ReactNode } from 'react';

function wrapper({ children }: { children: ReactNode }) {
  return <WorkspaceProvider>{children}</WorkspaceProvider>;
}

function useTestWorkspace() {
  return renderHook(() => useWorkspace(), { wrapper });
}

describe('WorkspaceContext', () => {
  it('provides initial state', () => {
    const { result } = useTestWorkspace();
    expect(result.current.state.sessionId).toBeNull();
    expect(result.current.state.files).toEqual([]);
    expect(result.current.state.openTabs).toEqual([]);
    expect(result.current.state.activeTabId).toBeNull();
    expect(result.current.state.activeActivity).toBe('explorer');
    expect(result.current.state.activePanel).toBe('terminal');
    expect(result.current.state.isPanelOpen).toBe(true);
    expect(result.current.state.isSidebarOpen).toBe(true);
    expect(result.current.state.isConnected).toBe(false);
  });

  it('throws when used outside provider', () => {
    expect(() => {
      renderHook(() => useWorkspace());
    }).toThrow('useWorkspace must be used within WorkspaceProvider');
  });

  it('SET_SESSION updates sessionId', () => {
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'SET_SESSION', sessionId: 'abc-123' }));
    expect(result.current.state.sessionId).toBe('abc-123');
  });

  it('SET_FILES updates files', () => {
    const files: FileNode[] = [
      { name: 'src', path: '/src', type: 'directory', children: [] },
      { name: 'main.ts', path: '/main.ts', type: 'file' },
    ];
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'SET_FILES', files }));
    expect(result.current.state.files).toEqual(files);
  });

  it('OPEN_FILE adds new tab and sets active', () => {
    const tab: EditorTab = {
      id: 'tab-1',
      path: '/main.ts',
      name: 'main.ts',
      language: 'typescript',
      content: '',
      isDirty: false,
    };
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab }));
    expect(result.current.state.openTabs).toHaveLength(1);
    expect(result.current.state.activeTabId).toBe('tab-1');
  });

  it('OPEN_FILE deduplicates existing tab', () => {
    const tab: EditorTab = {
      id: 'tab-1',
      path: '/main.ts',
      name: 'main.ts',
      language: 'typescript',
      content: '',
      isDirty: false,
    };
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab }));
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab }));
    expect(result.current.state.openTabs).toHaveLength(1);
    expect(result.current.state.activeTabId).toBe('tab-1');
  });

  it('CLOSE_TAB removes tab and selects last', () => {
    const tab1: EditorTab = { id: 't1', path: '/a', name: 'a', language: 'typescript', content: '', isDirty: false };
    const tab2: EditorTab = { id: 't2', path: '/b', name: 'b', language: 'typescript', content: '', isDirty: false };
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab: tab1 }));
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab: tab2 }));
    act(() => result.current.dispatch({ type: 'CLOSE_TAB', tabId: 't2' }));
    expect(result.current.state.openTabs).toHaveLength(1);
    expect(result.current.state.activeTabId).toBe('t1');
  });

  it('CLOSE_TAB sets null when last tab closed', () => {
    const tab: EditorTab = { id: 't1', path: '/a', name: 'a', language: 'typescript', content: '', isDirty: false };
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab }));
    act(() => result.current.dispatch({ type: 'CLOSE_TAB', tabId: 't1' }));
    expect(result.current.state.openTabs).toHaveLength(0);
    expect(result.current.state.activeTabId).toBeNull();
  });

  it('CLOSE_TAB does not change activeTabId when closing non-active tab', () => {
    const tab1: EditorTab = { id: 't1', path: '/a', name: 'a', language: 'typescript', content: '', isDirty: false };
    const tab2: EditorTab = { id: 't2', path: '/b', name: 'b', language: 'typescript', content: '', isDirty: false };
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab: tab1 }));
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab: tab2 }));
    // t2 is active. Close t1.
    act(() => result.current.dispatch({ type: 'CLOSE_TAB', tabId: 't1' }));
    expect(result.current.state.activeTabId).toBe('t2');
  });

  it('SET_ACTIVE_TAB changes active tab', () => {
    const tab1: EditorTab = { id: 't1', path: '/a', name: 'a', language: 'typescript', content: '', isDirty: false };
    const tab2: EditorTab = { id: 't2', path: '/b', name: 'b', language: 'typescript', content: '', isDirty: false };
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab: tab1 }));
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab: tab2 }));
    act(() => result.current.dispatch({ type: 'SET_ACTIVE_TAB', tabId: 't1' }));
    expect(result.current.state.activeTabId).toBe('t1');
  });

  it('UPDATE_TAB_CONTENT marks tab dirty', () => {
    const tab: EditorTab = { id: 't1', path: '/a', name: 'a', language: 'typescript', content: '', isDirty: false };
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab }));
    act(() => result.current.dispatch({ type: 'UPDATE_TAB_CONTENT', tabId: 't1', content: 'new content' }));
    expect(result.current.state.openTabs[0].content).toBe('new content');
    expect(result.current.state.openTabs[0].isDirty).toBe(true);
  });

  it('MARK_TAB_SAVED clears dirty flag', () => {
    const tab: EditorTab = { id: 't1', path: '/a', name: 'a', language: 'typescript', content: '', isDirty: false };
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'OPEN_FILE', tab }));
    act(() => result.current.dispatch({ type: 'UPDATE_TAB_CONTENT', tabId: 't1', content: 'edited' }));
    act(() => result.current.dispatch({ type: 'MARK_TAB_SAVED', tabId: 't1' }));
    expect(result.current.state.openTabs[0].isDirty).toBe(false);
  });

  it('SET_ACTIVITY changes activity and opens sidebar', () => {
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'SET_ACTIVITY', activity: 'search' }));
    expect(result.current.state.activeActivity).toBe('search');
    expect(result.current.state.isSidebarOpen).toBe(true);
  });

  it('SET_ACTIVITY toggles sidebar when same activity clicked', () => {
    const { result } = useTestWorkspace();
    // Explorer is default. Click explorer again.
    act(() => result.current.dispatch({ type: 'SET_ACTIVITY', activity: 'explorer' }));
    expect(result.current.state.isSidebarOpen).toBe(false);
    act(() => result.current.dispatch({ type: 'SET_ACTIVITY', activity: 'explorer' }));
    expect(result.current.state.isSidebarOpen).toBe(true);
  });

  it('SET_PANEL opens panel', () => {
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'TOGGLE_PANEL' })); // close
    act(() => result.current.dispatch({ type: 'SET_PANEL', panel: 'problems' }));
    expect(result.current.state.activePanel).toBe('problems');
    expect(result.current.state.isPanelOpen).toBe(true);
  });

  it('TOGGLE_PANEL flips isPanelOpen', () => {
    const { result } = useTestWorkspace();
    expect(result.current.state.isPanelOpen).toBe(true);
    act(() => result.current.dispatch({ type: 'TOGGLE_PANEL' }));
    expect(result.current.state.isPanelOpen).toBe(false);
    act(() => result.current.dispatch({ type: 'TOGGLE_PANEL' }));
    expect(result.current.state.isPanelOpen).toBe(true);
  });

  it('TOGGLE_SIDEBAR flips isSidebarOpen', () => {
    const { result } = useTestWorkspace();
    expect(result.current.state.isSidebarOpen).toBe(true);
    act(() => result.current.dispatch({ type: 'TOGGLE_SIDEBAR' }));
    expect(result.current.state.isSidebarOpen).toBe(false);
  });

  it('SET_CONNECTED updates isConnected', () => {
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'SET_CONNECTED', connected: true }));
    expect(result.current.state.isConnected).toBe(true);
    act(() => result.current.dispatch({ type: 'SET_CONNECTED', connected: false }));
    expect(result.current.state.isConnected).toBe(false);
  });

  it('TOGGLE_FILE_EXPAND toggles expansion', () => {
    const files: FileNode[] = [
      { name: 'src', path: '/src', type: 'directory', isExpanded: false, children: [] },
    ];
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'SET_FILES', files }));
    act(() => result.current.dispatch({ type: 'TOGGLE_FILE_EXPAND', path: '/src' }));
    expect(result.current.state.files[0].isExpanded).toBe(true);
    act(() => result.current.dispatch({ type: 'TOGGLE_FILE_EXPAND', path: '/src' }));
    expect(result.current.state.files[0].isExpanded).toBe(false);
  });

  it('TOGGLE_FILE_EXPAND works on nested nodes', () => {
    const files: FileNode[] = [
      {
        name: 'src',
        path: '/src',
        type: 'directory',
        isExpanded: true,
        children: [
          { name: 'lib', path: '/src/lib', type: 'directory', isExpanded: false, children: [] },
        ],
      },
    ];
    const { result } = useTestWorkspace();
    act(() => result.current.dispatch({ type: 'SET_FILES', files }));
    act(() => result.current.dispatch({ type: 'TOGGLE_FILE_EXPAND', path: '/src/lib' }));
    expect(result.current.state.files[0].children![0].isExpanded).toBe(true);
  });
});
