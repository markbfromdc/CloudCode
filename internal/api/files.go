package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

// FileOpsHandler provides create/delete operations for files and directories.
type FileOpsHandler struct {
	log      *logging.Logger
	basePath string
}

// NewFileOpsHandler creates a new handler for file create/delete operations.
func NewFileOpsHandler(log *logging.Logger) *FileOpsHandler {
	return &FileOpsHandler{
		log:      log.WithField("component", "fileops-api"),
		basePath: "/workspace",
	}
}

// HandleCreateFile creates a new file or directory.
func (h *FileOpsHandler) HandleCreateFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Path string `json:"path"`
		Type string `json:"type"` // "file" or "directory"
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if payload.Path == "" {
		http.Error(w, `{"error":"path required"}`, http.StatusBadRequest)
		return
	}

	cleaned := filepath.Clean(payload.Path)
	if !strings.HasPrefix(cleaned, h.basePath) {
		http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
		return
	}

	if payload.Type == "directory" {
		if err := os.MkdirAll(cleaned, 0755); err != nil {
			h.log.Error("failed to create directory %s: %v", cleaned, err)
			http.Error(w, `{"error":"failed to create directory"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// Ensure parent directory exists.
		if err := os.MkdirAll(filepath.Dir(cleaned), 0755); err != nil {
			h.log.Error("failed to create parent directory: %v", err)
			http.Error(w, `{"error":"failed to create parent directory"}`, http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(cleaned, []byte{}, 0644); err != nil {
			h.log.Error("failed to create file %s: %v", cleaned, err)
			http.Error(w, `{"error":"failed to create file"}`, http.StatusInternalServerError)
			return
		}
	}

	h.log.Info("created %s: %s", payload.Type, cleaned)
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "path": cleaned, "type": payload.Type})
}

// HandleDeleteFile deletes a file or directory.
func (h *FileOpsHandler) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, `{"error":"path required"}`, http.StatusBadRequest)
		return
	}

	cleaned := filepath.Clean(filePath)
	if !strings.HasPrefix(cleaned, h.basePath) {
		http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
		return
	}

	// Prevent deleting the workspace root.
	if cleaned == h.basePath {
		http.Error(w, `{"error":"cannot delete workspace root"}`, http.StatusForbidden)
		return
	}

	if err := os.RemoveAll(cleaned); err != nil {
		h.log.Error("failed to delete %s: %v", cleaned, err)
		http.Error(w, `{"error":"failed to delete"}`, http.StatusInternalServerError)
		return
	}

	h.log.Info("deleted: %s", cleaned)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "path": cleaned})
}

// HandleRenameFile renames/moves a file or directory.
func (h *FileOpsHandler) HandleRenameFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	oldClean := filepath.Clean(payload.OldPath)
	newClean := filepath.Clean(payload.NewPath)

	if !strings.HasPrefix(oldClean, h.basePath) || !strings.HasPrefix(newClean, h.basePath) {
		http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
		return
	}

	// Ensure parent directory of new path exists.
	if err := os.MkdirAll(filepath.Dir(newClean), 0755); err != nil {
		http.Error(w, `{"error":"failed to create parent directory"}`, http.StatusInternalServerError)
		return
	}

	if err := os.Rename(oldClean, newClean); err != nil {
		h.log.Error("failed to rename %s to %s: %v", oldClean, newClean, err)
		http.Error(w, `{"error":"failed to rename"}`, http.StatusInternalServerError)
		return
	}

	h.log.Info("renamed: %s -> %s", oldClean, newClean)
	writeJSON(w, http.StatusOK, map[string]string{"status": "renamed", "old_path": oldClean, "new_path": newClean})
}
