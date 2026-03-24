// Package api - git.go provides HTTP API handlers for Git operations
// within workspace containers, supporting commit, status, branch, and log.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

// GitStatus represents the status of a file in the working tree.
type GitStatus struct {
	Path       string `json:"path"`
	Status     string `json:"status"`
	StatusCode string `json:"status_code"`
}

// GitCommitInfo represents a single commit entry.
type GitCommitInfo struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	Message string `json:"message"`
}

// GitBranchInfo represents a Git branch.
type GitBranchInfo struct {
	Name    string `json:"name"`
	Current bool   `json:"current"`
}

// GitHandler serves Git operation endpoints for workspaces.
type GitHandler struct {
	log      *logging.Logger
	basePath string
}

// NewGitHandler creates a new handler for Git operations.
func NewGitHandler(log *logging.Logger) *GitHandler {
	return &GitHandler{
		log:      log.WithField("component", "git-api"),
		basePath: "/workspace",
	}
}

// HandleGitStatus returns the Git status of the workspace.
func (h *GitHandler) HandleGitStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	workDir := h.resolveWorkDir(r)
	if workDir == "" {
		http.Error(w, `{"error":"invalid workspace path"}`, http.StatusBadRequest)
		return
	}

	out, err := h.runGit(workDir, "status", "--porcelain", "-u")
	if err != nil {
		h.log.Error("git status failed in %s: %v", workDir, err)
		http.Error(w, `{"error":"git status failed"}`, http.StatusInternalServerError)
		return
	}

	statuses := parseGitStatus(out)
	writeJSON(w, http.StatusOK, statuses)
}

// HandleGitLog returns the recent commit history.
func (h *GitHandler) HandleGitLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	workDir := h.resolveWorkDir(r)
	if workDir == "" {
		http.Error(w, `{"error":"invalid workspace path"}`, http.StatusBadRequest)
		return
	}

	out, err := h.runGit(workDir, "log", "--oneline", "--format=%H|%an|%ad|%s", "--date=short", "-20")
	if err != nil {
		// Empty repo has no commits — not an error.
		writeJSON(w, http.StatusOK, []GitCommitInfo{})
		return
	}

	commits := parseGitLog(out)
	writeJSON(w, http.StatusOK, commits)
}

// HandleGitBranches returns the list of branches.
func (h *GitHandler) HandleGitBranches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	workDir := h.resolveWorkDir(r)
	if workDir == "" {
		http.Error(w, `{"error":"invalid workspace path"}`, http.StatusBadRequest)
		return
	}

	out, err := h.runGit(workDir, "branch", "--list", "--no-color")
	if err != nil {
		writeJSON(w, http.StatusOK, []GitBranchInfo{})
		return
	}

	branches := parseGitBranches(out)
	writeJSON(w, http.StatusOK, branches)
}

// HandleGitCommit creates a new commit with staged changes.
func (h *GitHandler) HandleGitCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	workDir := h.resolveWorkDir(r)
	if workDir == "" {
		http.Error(w, `{"error":"invalid workspace path"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		Message string   `json:"message"`
		Files   []string `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if payload.Message == "" {
		http.Error(w, `{"error":"commit message required"}`, http.StatusBadRequest)
		return
	}

	// Stage specified files, or all if none specified.
	if len(payload.Files) > 0 {
		for _, f := range payload.Files {
			clean := filepath.Clean(f)
			if _, err := h.runGit(workDir, "add", clean); err != nil {
				h.log.Error("git add failed for %s: %v", clean, err)
			}
		}
	} else {
		if _, err := h.runGit(workDir, "add", "-A"); err != nil {
			http.Error(w, `{"error":"git add failed"}`, http.StatusInternalServerError)
			return
		}
	}

	out, err := h.runGit(workDir, "commit", "--no-gpg-sign", "-m", payload.Message)
	if err != nil {
		h.log.Error("git commit failed: %v (output: %s)", err, out)
		http.Error(w, fmt.Sprintf(`{"error":"commit failed: %s"}`, strings.TrimSpace(out)), http.StatusInternalServerError)
		return
	}

	h.log.Info("git commit in %s: %s", workDir, payload.Message)
	writeJSON(w, http.StatusOK, map[string]string{"status": "committed", "output": strings.TrimSpace(out)})
}

// HandleGitStage stages files for the next commit.
func (h *GitHandler) HandleGitStage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	workDir := h.resolveWorkDir(r)
	if workDir == "" {
		http.Error(w, `{"error":"invalid workspace path"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		Files []string `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(payload.Files) == 0 {
		// Stage all changes.
		if _, err := h.runGit(workDir, "add", "-A"); err != nil {
			http.Error(w, `{"error":"git add failed"}`, http.StatusInternalServerError)
			return
		}
	} else {
		for _, f := range payload.Files {
			if _, err := h.runGit(workDir, "add", filepath.Clean(f)); err != nil {
				h.log.Warn("failed to stage %s: %v", f, err)
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "staged"})
}

// HandleGitInit initializes a new Git repository in the workspace.
func (h *GitHandler) HandleGitInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	workDir := h.resolveWorkDir(r)
	if workDir == "" {
		http.Error(w, `{"error":"invalid workspace path"}`, http.StatusBadRequest)
		return
	}

	if _, err := h.runGit(workDir, "init"); err != nil {
		http.Error(w, `{"error":"git init failed"}`, http.StatusInternalServerError)
		return
	}

	// Configure default user for commits inside container.
	h.runGit(workDir, "config", "user.email", "developer@cloudcode.dev")
	h.runGit(workDir, "config", "user.name", "CloudCode Developer")

	writeJSON(w, http.StatusOK, map[string]string{"status": "initialized"})
}

func (h *GitHandler) resolveWorkDir(r *http.Request) string {
	workDir := r.URL.Query().Get("workspace")
	if workDir == "" {
		workDir = h.basePath
	}
	cleaned := filepath.Clean(workDir)
	if !strings.HasPrefix(cleaned, h.basePath) {
		return ""
	}
	return cleaned
}

func (h *GitHandler) runGit(workDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func parseGitStatus(output string) []GitStatus {
	var statuses []GitStatus
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if len(line) < 4 {
			continue
		}
		code := strings.TrimSpace(line[:2])
		path := strings.TrimSpace(line[3:])
		status := "modified"
		switch {
		case strings.Contains(code, "A"):
			status = "added"
		case strings.Contains(code, "D"):
			status = "deleted"
		case strings.Contains(code, "R"):
			status = "renamed"
		case strings.Contains(code, "?"):
			status = "untracked"
		case strings.Contains(code, "M"):
			status = "modified"
		}
		statuses = append(statuses, GitStatus{
			Path:       path,
			Status:     status,
			StatusCode: code,
		})
	}
	return statuses
}

func parseGitLog(output string) []GitCommitInfo {
	var commits []GitCommitInfo
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}
		commits = append(commits, GitCommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    parts[2],
			Message: parts[3],
		})
	}
	return commits
}

func parseGitBranches(output string) []GitBranchInfo {
	var branches []GitBranchInfo
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		current := strings.HasPrefix(line, "* ")
		name := strings.TrimPrefix(strings.TrimSpace(line), "* ")
		branches = append(branches, GitBranchInfo{
			Name:    name,
			Current: current,
		})
	}
	return branches
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
