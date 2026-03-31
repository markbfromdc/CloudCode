package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

func setupTestWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "src", "components"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "main.ts"), []byte("console.log('hello');"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "components", "App.tsx"), []byte("<App />"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)

	return dir
}

func TestHandleListFiles(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	workDir := setupTestWorkspace(t)
	handler := NewFileTreeHandlerWithBase(log, workDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/test/files?path="+workDir, nil)
	rec := httptest.NewRecorder()

	handler.HandleListFiles(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var nodes []*FileNode
	if err := json.NewDecoder(rec.Body).Decode(&nodes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(nodes) < 2 {
		t.Fatalf("expected at least 2 nodes, got %d", len(nodes))
	}

	// First entry should be the 'src' directory (dirs sort first).
	if nodes[0].Name != "src" || nodes[0].Type != "directory" {
		t.Errorf("expected first node to be 'src' directory, got %s (%s)", nodes[0].Name, nodes[0].Type)
	}
}

func TestHandleListFilesMethodNotAllowed(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewFileTreeHandlerWithBase(log, "/tmp")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/test/files?path=/tmp", nil)
	rec := httptest.NewRecorder()

	handler.HandleListFiles(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleReadFile(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	workDir := setupTestWorkspace(t)
	handler := NewFileTreeHandlerWithBase(log, workDir)

	filePath := filepath.Join(workDir, "package.json")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/test/files/content?path="+filePath, nil)
	rec := httptest.NewRecorder()

	handler.HandleReadFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if body != `{"name":"test"}` {
		t.Errorf("expected file content, got %q", body)
	}
}

func TestHandleReadFileNotFound(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	workDir := setupTestWorkspace(t)
	handler := NewFileTreeHandlerWithBase(log, workDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/test/files/content?path="+filepath.Join(workDir, "nonexistent.txt"), nil)
	rec := httptest.NewRecorder()

	handler.HandleReadFile(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleWriteFile(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	workDir := setupTestWorkspace(t)
	handler := NewFileTreeHandlerWithBase(log, workDir)

	filePath := filepath.Join(workDir, "newfile.txt")

	body := strings.NewReader(`{"content":"hello world"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/test/files/content?path="+filePath, body)
	rec := httptest.NewRecorder()

	handler.HandleWriteFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
}

func TestHandleWriteFileMissingPath(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewFileTreeHandlerWithBase(log, "/tmp")

	req := httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/test/files/content", nil)
	rec := httptest.NewRecorder()

	handler.HandleWriteFile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleReadFileAccessDenied(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := NewFileTreeHandlerWithBase(log, "/workspace")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/test/files/content?path=/etc/passwd", nil)
	rec := httptest.NewRecorder()

	handler.HandleReadFile(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for path traversal, got %d", rec.Code)
	}
}

func TestBuildFileTreeSkipsHidden(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("ok"), 0644)
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte("hidden"), 0644)
	os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0755)
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	nodes, err := buildFileTree(dir, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, n := range nodes {
		if strings.HasPrefix(n.Name, ".") {
			t.Errorf("hidden file %q should be filtered", n.Name)
		}
		if n.Name == "node_modules" {
			t.Error("node_modules should be filtered")
		}
	}
}
