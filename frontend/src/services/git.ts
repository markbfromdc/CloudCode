const API_BASE = '/api/v1';

function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem('cloudcode_token');
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

export interface GitFileStatus {
  path: string;
  status: string;
  status_code: string;
}

export interface GitCommit {
  hash: string;
  author: string;
  date: string;
  message: string;
}

export interface GitBranch {
  name: string;
  current: boolean;
}

/** Get the Git status of the workspace. */
export async function getGitStatus(workspace?: string): Promise<GitFileStatus[]> {
  const params = workspace ? `?workspace=${encodeURIComponent(workspace)}` : '';
  const res = await fetch(`${API_BASE}/git/status${params}`, { headers: getAuthHeaders() });
  if (!res.ok) return [];
  return res.json();
}

/** Get the Git commit log. */
export async function getGitLog(workspace?: string): Promise<GitCommit[]> {
  const params = workspace ? `?workspace=${encodeURIComponent(workspace)}` : '';
  const res = await fetch(`${API_BASE}/git/log${params}`, { headers: getAuthHeaders() });
  if (!res.ok) return [];
  return res.json();
}

/** Get the list of Git branches. */
export async function getGitBranches(workspace?: string): Promise<GitBranch[]> {
  const params = workspace ? `?workspace=${encodeURIComponent(workspace)}` : '';
  const res = await fetch(`${API_BASE}/git/branches${params}`, { headers: getAuthHeaders() });
  if (!res.ok) return [];
  return res.json();
}

/** Stage files for commit. */
export async function stageFiles(files: string[], workspace?: string): Promise<void> {
  const params = workspace ? `?workspace=${encodeURIComponent(workspace)}` : '';
  await fetch(`${API_BASE}/git/stage${params}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
    body: JSON.stringify({ files }),
  });
}

/** Create a Git commit. */
export async function createCommit(
  message: string,
  files?: string[],
  workspace?: string
): Promise<{ status: string; output: string }> {
  const params = workspace ? `?workspace=${encodeURIComponent(workspace)}` : '';
  const res = await fetch(`${API_BASE}/git/commit${params}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
    body: JSON.stringify({ message, files }),
  });
  if (!res.ok) throw new Error('Commit failed');
  return res.json();
}

/** Initialize a new Git repository. */
export async function initRepo(workspace?: string): Promise<void> {
  const params = workspace ? `?workspace=${encodeURIComponent(workspace)}` : '';
  await fetch(`${API_BASE}/git/init${params}`, { method: 'POST', headers: getAuthHeaders() });
}
