import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getGitStatus, getGitLog, getGitBranches, stageFiles, createCommit, initRepo } from './git';

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
  });
}

describe('Git service', () => {
  describe('getGitStatus', () => {
    it('fetches git status', async () => {
      const statuses = [{ path: 'main.ts', status: 'modified', status_code: 'M' }];
      mockFetch.mockReturnValue(jsonResponse(statuses));

      const result = await getGitStatus();

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/git/status');
      expect(result).toEqual(statuses);
    });

    it('passes workspace parameter', async () => {
      mockFetch.mockReturnValue(jsonResponse([]));

      await getGitStatus('/workspace/project');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/git/status?workspace=%2Fworkspace%2Fproject'
      );
    });

    it('returns empty array on error', async () => {
      mockFetch.mockReturnValue(jsonResponse(null, 500));

      const result = await getGitStatus();
      expect(result).toEqual([]);
    });
  });

  describe('getGitLog', () => {
    it('fetches commit log', async () => {
      const commits = [{ hash: 'abc', author: 'Dev', date: '2024-01-01', message: 'init' }];
      mockFetch.mockReturnValue(jsonResponse(commits));

      const result = await getGitLog();
      expect(result).toEqual(commits);
    });

    it('returns empty array on error', async () => {
      mockFetch.mockReturnValue(jsonResponse(null, 500));

      const result = await getGitLog();
      expect(result).toEqual([]);
    });
  });

  describe('getGitBranches', () => {
    it('fetches branches', async () => {
      const branches = [{ name: 'main', current: true }, { name: 'dev', current: false }];
      mockFetch.mockReturnValue(jsonResponse(branches));

      const result = await getGitBranches();
      expect(result).toEqual(branches);
    });

    it('returns empty array on error', async () => {
      mockFetch.mockReturnValue(jsonResponse(null, 404));

      const result = await getGitBranches();
      expect(result).toEqual([]);
    });
  });

  describe('stageFiles', () => {
    it('sends POST with files', async () => {
      mockFetch.mockReturnValue(jsonResponse({}));

      await stageFiles(['main.ts', 'app.ts']);

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/git/stage', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ files: ['main.ts', 'app.ts'] }),
      }));
    });

    it('passes workspace parameter', async () => {
      mockFetch.mockReturnValue(jsonResponse({}));

      await stageFiles([], '/workspace');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/git/stage?workspace=%2Fworkspace',
        expect.anything()
      );
    });
  });

  describe('createCommit', () => {
    it('sends POST with message', async () => {
      mockFetch.mockReturnValue(jsonResponse({ status: 'committed', output: 'ok' }));

      const result = await createCommit('fix bug');

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/git/commit', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ message: 'fix bug' }),
      }));
      expect(result.status).toBe('committed');
    });

    it('includes optional files', async () => {
      mockFetch.mockReturnValue(jsonResponse({ status: 'committed', output: 'ok' }));

      await createCommit('update', ['a.ts']);

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/git/commit', expect.objectContaining({
        body: JSON.stringify({ message: 'update', files: ['a.ts'] }),
      }));
    });

    it('throws on error', async () => {
      mockFetch.mockReturnValue(jsonResponse(null, 500));

      await expect(createCommit('msg')).rejects.toThrow('Commit failed');
    });
  });

  describe('initRepo', () => {
    it('sends POST to init', async () => {
      mockFetch.mockReturnValue(jsonResponse({}));

      await initRepo();

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/git/init', expect.objectContaining({
        method: 'POST',
      }));
    });

    it('passes workspace parameter', async () => {
      mockFetch.mockReturnValue(jsonResponse({}));

      await initRepo('/workspace');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/git/init?workspace=%2Fworkspace',
        expect.objectContaining({ method: 'POST' })
      );
    });
  });
});
