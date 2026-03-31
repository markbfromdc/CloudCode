// Package api provides HTTP API handlers for workspace operations
// including file tree browsing, file reading, and file writing.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

// FileNode represents a file or directory in the workspace file tree.
type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     string      `json:"type"`
	Children []*FileNode `json:"children,omitempty"`
}

// FileTreeHandler serves the workspace file tree and file operations.
type FileTreeHandler struct {
	log      *logging.Logger
	basePath string
}

// NewFileTreeHandler creates a new handler for file tree operations.
// The basePath restricts file access to a specific directory prefix for security.
func NewFileTreeHandler(log *logging.Logger) *FileTreeHandler {
	return &FileTreeHandler{
		log:      log.WithField("component", "filetree-api"),
		basePath: "/workspace",
	}
}

// NewFileTreeHandlerWithBase creates a handler with a custom base path (for testing).
func NewFileTreeHandlerWithBase(log *logging.Logger, basePath string) *FileTreeHandler {
	return &FileTreeHandler{
		log:      log.WithField("component", "filetree-api"),
		basePath: basePath,
	}
}

// HandleListFiles returns the file tree for a workspace directory.
func (h *FileTreeHandler) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	dirPath := r.URL.Query().Get("path")
	if dirPath == "" {
		dirPath = h.basePath
	}

	// Prevent directory traversal attacks.
	cleaned := filepath.Clean(dirPath)
	if !strings.HasPrefix(cleaned, h.basePath) {
		http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
		return
	}

	nodes, err := buildFileTree(cleaned, 3)
	if err != nil {
		h.log.Error("failed to build file tree for %s: %v", cleaned, err)
		http.Error(w, `{"error":"failed to read directory"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// HandleReadFile returns the contents of a file.
func (h *FileTreeHandler) HandleReadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	info, err := os.Stat(cleaned)
	if err != nil {
		http.Error(w, `{"error":"file not found"}`, http.StatusNotFound)
		return
	}

	// Limit file size to 10MB.
	if info.Size() > 10*1024*1024 {
		http.Error(w, `{"error":"file too large"}`, http.StatusRequestEntityTooLarge)
		return
	}

	data, err := os.ReadFile(cleaned)
	if err != nil {
		h.log.Error("failed to read file %s: %v", cleaned, err)
		http.Error(w, `{"error":"failed to read file"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

// HandleWriteFile writes content to a file in the workspace.
func (h *FileTreeHandler) HandleWriteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	// Limit request body to 10MB.
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	var payload struct {
		Content string `json:"content"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"request body too large"}`, http.StatusRequestEntityTooLarge)
		return
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Ensure parent directory exists.
	dir := filepath.Dir(cleaned)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.log.Error("failed to create directory %s: %v", dir, err)
		http.Error(w, `{"error":"failed to create directory"}`, http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(cleaned, []byte(payload.Content), 0644); err != nil {
		h.log.Error("failed to write file %s: %v", cleaned, err)
		http.Error(w, `{"error":"failed to write file"}`, http.StatusInternalServerError)
		return
	}

	h.log.Info("file written: %s (%d bytes)", cleaned, len(payload.Content))
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","path":"%s","size":%d}`, cleaned, len(payload.Content))
}

// buildFileTree recursively builds a file tree from a directory path up to maxDepth.
func buildFileTree(rootPath string, maxDepth int) ([]*FileNode, error) {
	return readDir(rootPath, 0, maxDepth)
}

func readDir(dirPath string, depth, maxDepth int) ([]*FileNode, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var nodes []*FileNode

	// Sort: directories first, then files, alphabetically.
	sort.Slice(entries, func(i, j int) bool {
		iDir := entries[i].IsDir()
		jDir := entries[j].IsDir()
		if iDir != jDir {
			return iDir
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files and common non-essential directories.
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "__pycache__" || name == "vendor" {
			continue
		}

		fullPath := filepath.Join(dirPath, name)

		node := &FileNode{
			Name: name,
			Path: fullPath,
		}

		if entry.IsDir() {
			node.Type = "directory"
			if depth < maxDepth {
				children, err := readDir(fullPath, depth+1, maxDepth)
				if err == nil {
					node.Children = children
				}
			}
		} else {
			node.Type = "file"
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
