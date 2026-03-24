import { createContext, useContext, useReducer, type ReactNode } from 'react';
import type { EditorTab, FileNode, ActivityBarItem, PanelTab } from '../types';

interface WorkspaceState {
  sessionId: string | null;
  files: FileNode[];
  openTabs: EditorTab[];
  activeTabId: string | null;
  activeActivity: ActivityBarItem;
  activePanel: PanelTab;
  isPanelOpen: boolean;
  isSidebarOpen: boolean;
  isConnected: boolean;
}

type Action =
  | { type: 'SET_SESSION'; sessionId: string }
  | { type: 'SET_FILES'; files: FileNode[] }
  | { type: 'OPEN_FILE'; tab: EditorTab }
  | { type: 'CLOSE_TAB'; tabId: string }
  | { type: 'SET_ACTIVE_TAB'; tabId: string }
  | { type: 'UPDATE_TAB_CONTENT'; tabId: string; content: string }
  | { type: 'MARK_TAB_SAVED'; tabId: string }
  | { type: 'SET_ACTIVITY'; activity: ActivityBarItem }
  | { type: 'SET_PANEL'; panel: PanelTab }
  | { type: 'TOGGLE_PANEL' }
  | { type: 'TOGGLE_SIDEBAR' }
  | { type: 'SET_CONNECTED'; connected: boolean }
  | { type: 'TOGGLE_FILE_EXPAND'; path: string };

const initialState: WorkspaceState = {
  sessionId: null,
  files: [],
  openTabs: [],
  activeTabId: null,
  activeActivity: 'explorer',
  activePanel: 'terminal',
  isPanelOpen: true,
  isSidebarOpen: true,
  isConnected: false,
};

function toggleExpand(nodes: FileNode[], path: string): FileNode[] {
  return nodes.map((node) => {
    if (node.path === path) {
      return { ...node, isExpanded: !node.isExpanded };
    }
    if (node.children) {
      return { ...node, children: toggleExpand(node.children, path) };
    }
    return node;
  });
}

function reducer(state: WorkspaceState, action: Action): WorkspaceState {
  switch (action.type) {
    case 'SET_SESSION':
      return { ...state, sessionId: action.sessionId };

    case 'SET_FILES':
      return { ...state, files: action.files };

    case 'OPEN_FILE': {
      const existing = state.openTabs.find((t) => t.id === action.tab.id);
      if (existing) {
        return { ...state, activeTabId: existing.id };
      }
      return {
        ...state,
        openTabs: [...state.openTabs, action.tab],
        activeTabId: action.tab.id,
      };
    }

    case 'CLOSE_TAB': {
      const tabs = state.openTabs.filter((t) => t.id !== action.tabId);
      let activeTabId = state.activeTabId;
      if (activeTabId === action.tabId) {
        activeTabId = tabs.length > 0 ? tabs[tabs.length - 1].id : null;
      }
      return { ...state, openTabs: tabs, activeTabId };
    }

    case 'SET_ACTIVE_TAB':
      return { ...state, activeTabId: action.tabId };

    case 'UPDATE_TAB_CONTENT':
      return {
        ...state,
        openTabs: state.openTabs.map((t) =>
          t.id === action.tabId ? { ...t, content: action.content, isDirty: true } : t
        ),
      };

    case 'MARK_TAB_SAVED':
      return {
        ...state,
        openTabs: state.openTabs.map((t) =>
          t.id === action.tabId ? { ...t, isDirty: false } : t
        ),
      };

    case 'SET_ACTIVITY':
      return {
        ...state,
        activeActivity: action.activity,
        isSidebarOpen: state.activeActivity === action.activity ? !state.isSidebarOpen : true,
      };

    case 'SET_PANEL':
      return { ...state, activePanel: action.panel, isPanelOpen: true };

    case 'TOGGLE_PANEL':
      return { ...state, isPanelOpen: !state.isPanelOpen };

    case 'TOGGLE_SIDEBAR':
      return { ...state, isSidebarOpen: !state.isSidebarOpen };

    case 'SET_CONNECTED':
      return { ...state, isConnected: action.connected };

    case 'TOGGLE_FILE_EXPAND':
      return { ...state, files: toggleExpand(state.files, action.path) };

    default:
      return state;
  }
}

const WorkspaceContext = createContext<{
  state: WorkspaceState;
  dispatch: React.Dispatch<Action>;
} | null>(null);

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState);
  return (
    <WorkspaceContext.Provider value={{ state, dispatch }}>
      {children}
    </WorkspaceContext.Provider>
  );
}

export function useWorkspace() {
  const ctx = useContext(WorkspaceContext);
  if (!ctx) throw new Error('useWorkspace must be used within WorkspaceProvider');
  return ctx;
}
