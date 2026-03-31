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

func newFileOpsHandler(t *testing.T) (*FileOpsHandler, string) {
	t.Helper()
	dir := t.TempDir()
	log := logging.New(nil, logging.INFO)
	handler := &FileOpsHandler{log: log, basePath: dir}
	return handler, dir
}

func TestHandleCreateFile(t *testing.T) {
	handler, dir := newFileOpsHandler(t)
	filePath := filepath.Join(dir, "newfile.txt")

	body := strings.NewReader(`{"path":"` + filePath + `","type":"file"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/create", body)
	rec := httptest.NewRecorder()

	handler.HandleCreateFile(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}

func TestHandleCreateDirectory(t *testing.T) {
	handler, dir := newFileOpsHandler(t)
	dirPath := filepath.Join(dir, "subdir", "nested")

	body := strings.NewReader(`{"path":"` + dirPath + `","type":"directory"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/create", body)
	rec := httptest.NewRecorder()

	handler.HandleCreateFile(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
	if !info.IsDir() {
		t.Error("expected a directory, got file")
	}
}

func TestHandleDeleteFile(t *testing.T) {
	handler, dir := newFileOpsHandler(t)
	filePath := filepath.Join(dir, "deleteme.txt")
	os.WriteFile(filePath, []byte("delete"), 0644)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/delete?path="+filePath, nil)
	rec := httptest.NewRecorder()

	handler.HandleDeleteFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}
}

func TestHandleDeleteFileAccessDenied(t *testing.T) {
	handler, _ := newFileOpsHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/delete?path=/etc/passwd", nil)
	rec := httptest.NewRecorder()

	handler.HandleDeleteFile(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestHandleDeleteWorkspaceRoot(t *testing.T) {
	handler, dir := newFileOpsHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/delete?path="+dir, nil)
	rec := httptest.NewRecorder()

	handler.HandleDeleteFile(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for root deletion, got %d", rec.Code)
	}
}

func TestHandleRenameFile(t *testing.T) {
	handler, dir := newFileOpsHandler(t)
	oldPath := filepath.Join(dir, "old.txt")
	newPath := filepath.Join(dir, "new.txt")
	os.WriteFile(oldPath, []byte("content"), 0644)

	body := strings.NewReader(`{"old_path":"` + oldPath + `","new_path":"` + newPath + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/rename", body)
	rec := httptest.NewRecorder()

	handler.HandleRenameFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("expected old file to not exist")
	}
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("expected new file to exist")
	}
}

func TestHandleRenameFileAccessDenied(t *testing.T) {
	handler, dir := newFileOpsHandler(t)

	body := strings.NewReader(`{"old_path":"` + filepath.Join(dir, "a.txt") + `","new_path":"/etc/evil"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/rename", body)
	rec := httptest.NewRecorder()

	handler.HandleRenameFile(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
