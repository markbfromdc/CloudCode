/** Represents a file or directory in the workspace file tree. */
export interface FileNode {
  name: string;
  path: string;
  type: 'file' | 'directory';
  children?: FileNode[];
  isExpanded?: boolean;
}

/** Represents an open editor tab. */
export interface EditorTab {
  id: string;
  path: string;
  name: string;
  language: string;
  content: string;
  isDirty: boolean;
}

/** Workspace session information returned by the backend. */
export interface WorkspaceSession {
  session_id: string;
  container_id: string;
  status: string;
}

/** Backend health check response. */
export interface HealthStatus {
  status: string;
  active_sessions: number;
  active_workspaces: number;
  timestamp: string;
}

/** WebSocket message types for terminal communication. */
export type WSMessageType = 'input' | 'output' | 'resize' | 'heartbeat';

export interface WSMessage {
  type: WSMessageType;
  data?: string;
  cols?: number;
  rows?: number;
}

/** Activity bar item identifiers. */
export type ActivityBarItem = 'explorer' | 'search' | 'git' | 'extensions' | 'settings';

/** Panel identifiers for the bottom panel area. */
export type PanelTab = 'terminal' | 'output' | 'problems';
