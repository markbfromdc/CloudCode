package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

func TestHandleReadFileTooLarge(t *testing.T) {
	dir := t.TempDir()
	log := logging.New(nil, logging.INFO)
	handler := NewFileTreeHandlerWithBase(log, dir)

	// Create a file larger than 10MB.
	bigFile := filepath.Join(dir, "big.bin")
	data := make([]byte, 11*1024*1024) // 11MB
	os.WriteFile(bigFile, data, 0644)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/s1/files/content?path="+bigFile, nil)
	rec := httptest.NewRecorder()
	handler.HandleReadFile(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for large file, got %d", rec.Code)
	}
}

func TestHandleWriteFileEmptyContent(t *testing.T) {
	dir := t.TempDir()
	log := logging.New(nil, logging.INFO)
	handler := NewFileTreeHandlerWithBase(log, dir)

	filePath := filepath.Join(dir, "empty.txt")
	body := strings.NewReader(`{"content":""}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/s1/files/content?path="+filePath, body)
	rec := httptest.NewRecorder()
	handler.HandleWriteFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	content, _ := os.ReadFile(filePath)
	if string(content) != "" {
		t.Errorf("expected empty content, got %q", string(content))
	}
}

func TestHandleListFilesEmpty(t *testing.T) {
	dir := t.TempDir()
	log := logging.New(nil, logging.INFO)
	handler := NewFileTreeHandlerWithBase(log, dir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/s1/files?path="+dir, nil)
	rec := httptest.NewRecorder()
	handler.HandleListFiles(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	// Empty directory returns null (Go nil slice) or [].
	if body != "[]" && body != "null" {
		t.Errorf("expected empty result, got: %s", body)
	}
}

func TestHandleGitStatusMethodNotAllowed(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewGitHandler(log)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/status", nil)
	rec := httptest.NewRecorder()
	handler.HandleGitStatus(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleGitLogMethodNotAllowed(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewGitHandler(log)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/git/log", nil)
	rec := httptest.NewRecorder()
	handler.HandleGitLog(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleGitBranchesMethodNotAllowed(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewGitHandler(log)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/git/branches", nil)
	rec := httptest.NewRecorder()
	handler.HandleGitBranches(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleGitCommitMethodNotAllowed(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewGitHandler(log)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/git/commit", nil)
	rec := httptest.NewRecorder()
	handler.HandleGitCommit(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleGitStageMethodNotAllowed(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewGitHandler(log)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/git/stage", nil)
	rec := httptest.NewRecorder()
	handler.HandleGitStage(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleGitInitMethodNotAllowed(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewGitHandler(log)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/git/init", nil)
	rec := httptest.NewRecorder()
	handler.HandleGitInit(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleGitStageAll(t *testing.T) {
	dir, handler := setupGitWorkspace(t)

	// Create files.
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)

	// Stage all (empty files array).
	body := strings.NewReader(`{"files":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/stage?workspace="+dir, body)
	rec := httptest.NewRecorder()
	handler.HandleGitStage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleGitCommitInvalidJSON(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := &GitHandler{log: log, basePath: "/workspace"}

	body := strings.NewReader(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/commit?workspace=/workspace", body)
	rec := httptest.NewRecorder()
	handler.HandleGitCommit(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestHandleGitStageInvalidJSON(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := &GitHandler{log: log, basePath: "/workspace"}

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/stage?workspace=/workspace", body)
	rec := httptest.NewRecorder()
	handler.HandleGitStage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestResolveWorkDirSecurity(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := &GitHandler{log: log, basePath: "/workspace"}

	// Valid workspace path.
	req := httptest.NewRequest(http.MethodGet, "/?workspace=/workspace/project", nil)
	if handler.resolveWorkDir(req) != "/workspace/project" {
		t.Error("expected /workspace/project")
	}

	// Path traversal attempt.
	req = httptest.NewRequest(http.MethodGet, "/?workspace=/etc/passwd", nil)
	if handler.resolveWorkDir(req) != "" {
		t.Error("expected empty string for path traversal")
	}

	// Default path.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	if handler.resolveWorkDir(req) != "/workspace" {
		t.Error("expected default /workspace")
	}
}

func TestHandleCreateFileInvalidJSON(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/create", body)
	rec := httptest.NewRecorder()
	handler.HandleCreateFile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleCreateFileEmptyPath(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	body := strings.NewReader(`{"path":"","type":"file"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/create", body)
	rec := httptest.NewRecorder()
	handler.HandleCreateFile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleCreateFileAccessDenied(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	body := strings.NewReader(`{"path":"/etc/evil","type":"file"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/create", body)
	rec := httptest.NewRecorder()
	handler.HandleCreateFile(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestHandleCreateFileMethodNotAllowed(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/create", nil)
	rec := httptest.NewRecorder()
	handler.HandleCreateFile(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleDeleteFileMethodNotAllowed(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/delete", nil)
	rec := httptest.NewRecorder()
	handler.HandleDeleteFile(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleDeleteFileMissingPath(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/delete", nil)
	rec := httptest.NewRecorder()
	handler.HandleDeleteFile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleRenameFileMethodNotAllowed(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/rename", nil)
	rec := httptest.NewRecorder()
	handler.HandleRenameFile(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleRenameFileInvalidJSON(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/rename", body)
	rec := httptest.NewRecorder()
	handler.HandleRenameFile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
