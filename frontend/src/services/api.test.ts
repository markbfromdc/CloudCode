import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createWorkspace, stopWorkspace, getHealth, listFiles, readFile, writeFile } from './api';

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

beforeEach(() => {
  mockFetch.mockReset();
});

function jsonResponse(data: unknown, status = 200) {
  return Promise.resolve({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
  });
}

describe('API service', () => {
  describe('createWorkspace', () => {
    it('sends POST to /api/v1/workspaces', async () => {
      const session = { session_id: 's1', container_id: 'c1', status: 'running' };
      mockFetch.mockReturnValue(jsonResponse(session));

      const result = await createWorkspace();

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/workspaces', expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }));
      expect(result).toEqual(session);
    });
  });

  describe('stopWorkspace', () => {
    it('sends POST with session_id', async () => {
      mockFetch.mockReturnValue(jsonResponse({}));

      await stopWorkspace('session-abc');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/workspaces/stop?session_id=session-abc',
        expect.objectContaining({ method: 'POST' })
      );
    });

    it('encodes special characters', async () => {
      mockFetch.mockReturnValue(jsonResponse({}));

      await stopWorkspace('session with spaces');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('session%20with%20spaces'),
        expect.anything()
      );
    });
  });

  describe('getHealth', () => {
    it('fetches /health endpoint', async () => {
      const health = { status: 'healthy', active_sessions: 0, active_workspaces: 0, timestamp: '2024-01-01' };
      mockFetch.mockReturnValue(jsonResponse(health));

      const result = await getHealth();

      expect(mockFetch).toHaveBeenCalledWith('/health');
      expect(result).toEqual(health);
    });
  });

  describe('listFiles', () => {
    it('fetches file list for session', async () => {
      const files = [{ name: 'main.ts', path: '/main.ts', type: 'file' }];
      mockFetch.mockReturnValue(jsonResponse(files));

      const result = await listFiles('s1', '/src');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/workspaces/s1/files?path=%2Fsrc',
        expect.anything()
      );
      expect(result).toEqual(files);
    });

    it('defaults to root path', async () => {
      mockFetch.mockReturnValue(jsonResponse([]));

      await listFiles('s1');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/workspaces/s1/files?path=%2F',
        expect.anything()
      );
    });
  });

  describe('readFile', () => {
    it('fetches file content as text', async () => {
      mockFetch.mockReturnValue(Promise.resolve({
        ok: true,
        status: 200,
        text: () => Promise.resolve('file content here'),
      }));

      const result = await readFile('s1', '/main.ts');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/workspaces/s1/files/content?path=%2Fmain.ts',
        expect.objectContaining({ headers: expect.any(Object) })
      );
      expect(result).toBe('file content here');
    });

    it('throws on error', async () => {
      mockFetch.mockReturnValue(Promise.resolve({
        ok: false,
        status: 404,
      }));

      await expect(readFile('s1', '/missing')).rejects.toThrow('Failed to read file: 404');
    });
  });

  describe('writeFile', () => {
    it('sends PUT with content', async () => {
      mockFetch.mockReturnValue(jsonResponse({ status: 'ok' }));

      await writeFile('s1', '/main.ts', 'new content');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/workspaces/s1/files/content?path=%2Fmain.ts',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ content: 'new content' }),
        })
      );
    });
  });

  describe('error handling', () => {
    it('throws on non-ok response', async () => {
      mockFetch.mockReturnValue(Promise.resolve({
        ok: false,
        status: 500,
        text: () => Promise.resolve('internal server error'),
      }));

      await expect(createWorkspace()).rejects.toThrow('API error 500: internal server error');
    });
  });
});
