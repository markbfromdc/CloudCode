import type { WorkspaceSession, HealthStatus, FileNode } from '../types';

const API_BASE = '/api/v1';

/** Generic fetch wrapper with error handling. */
async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('cloudcode_token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options?.headers as Record<string, string>,
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API error ${res.status}: ${body}`);
  }

  return res.json();
}

/** Create a new workspace container. */
export async function createWorkspace(): Promise<WorkspaceSession> {
  return request<WorkspaceSession>('/workspaces', { method: 'POST' });
}

/** Stop a workspace container. */
export async function stopWorkspace(sessionId: string): Promise<void> {
  await request(`/workspaces/stop?session_id=${encodeURIComponent(sessionId)}`, {
    method: 'POST',
  });
}

/** Get server health status. */
export async function getHealth(): Promise<HealthStatus> {
  const res = await fetch('/health');
  return res.json();
}

/** List files in a workspace directory. */
export async function listFiles(sessionId: string, dirPath: string = '/'): Promise<FileNode[]> {
  return request<FileNode[]>(`/workspaces/${sessionId}/files?path=${encodeURIComponent(dirPath)}`);
}

/** Read file content from workspace. */
export async function readFile(sessionId: string, filePath: string): Promise<string> {
  const token = localStorage.getItem('cloudcode_token');
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const res = await fetch(`${API_BASE}/workspaces/${sessionId}/files/content?path=${encodeURIComponent(filePath)}`, { headers });
  if (!res.ok) throw new Error(`Failed to read file: ${res.status}`);
  return res.text();
}

/** Write file content to workspace. */
export async function writeFile(sessionId: string, filePath: string, content: string): Promise<void> {
  await request(`/workspaces/${sessionId}/files/content?path=${encodeURIComponent(filePath)}`, {
    method: 'PUT',
    body: JSON.stringify({ content }),
  });
}
