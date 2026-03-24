package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

func setupGitWorkspace(t *testing.T) (string, *GitHandler) {
	t.Helper()
	dir := t.TempDir()

	// Initialize a git repo with signing disabled for test isolation.
	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.email", "test@test.com")
	runCmd(t, dir, "git", "config", "user.name", "Test")
	runCmd(t, dir, "git", "config", "commit.gpgsign", "false")
	runCmd(t, dir, "git", "config", "gpg.format", "openpgp")

	// Create a file and initial commit.
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello"), 0644)
	runCmd(t, dir, "git", "add", "-A")
	runCmd(t, dir, "git", "commit", "--no-gpg-sign", "-m", "initial commit")

	log := logging.New(nil, logging.INFO)
	handler := &GitHandler{log: log, basePath: dir}
	return dir, handler
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@test.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func TestHandleGitStatus(t *testing.T) {
	dir, handler := setupGitWorkspace(t)

	// Create an untracked file.
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/git/status?workspace="+dir, nil)
	rec := httptest.NewRecorder()
	handler.HandleGitStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "new.txt") {
		t.Errorf("expected new.txt in status output, got: %s", body)
	}
	if !strings.Contains(body, "untracked") {
		t.Errorf("expected untracked status, got: %s", body)
	}
}

func TestHandleGitLog(t *testing.T) {
	_, handler := setupGitWorkspace(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/git/log?workspace="+handler.basePath, nil)
	rec := httptest.NewRecorder()
	handler.HandleGitLog(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "initial commit") {
		t.Errorf("expected 'initial commit' in log, got: %s", body)
	}
}

func TestHandleGitBranches(t *testing.T) {
	_, handler := setupGitWorkspace(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/git/branches?workspace="+handler.basePath, nil)
	rec := httptest.NewRecorder()
	handler.HandleGitBranches(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "true") {
		t.Errorf("expected a current branch, got: %s", body)
	}
}

func TestHandleGitCommit(t *testing.T) {
	dir, handler := setupGitWorkspace(t)

	// Create a new file to commit.
	os.WriteFile(filepath.Join(dir, "committed.txt"), []byte("data"), 0644)

	body := strings.NewReader(`{"message":"test commit","files":["committed.txt"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/commit?workspace="+dir, body)
	rec := httptest.NewRecorder()
	handler.HandleGitCommit(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "committed") {
		t.Errorf("expected committed status, got: %s", rec.Body.String())
	}
}

func TestHandleGitCommitEmptyMessage(t *testing.T) {
	_, handler := setupGitWorkspace(t)

	body := strings.NewReader(`{"message":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/commit?workspace="+handler.basePath, body)
	rec := httptest.NewRecorder()
	handler.HandleGitCommit(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleGitInit(t *testing.T) {
	dir := t.TempDir()
	log := logging.New(nil, logging.INFO)
	handler := &GitHandler{log: log, basePath: dir}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/init?workspace="+dir, nil)
	rec := httptest.NewRecorder()
	handler.HandleGitInit(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify .git directory was created.
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		t.Error("expected .git directory to be created")
	}
}

func TestHandleGitStage(t *testing.T) {
	dir, handler := setupGitWorkspace(t)

	os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("data"), 0644)

	body := strings.NewReader(`{"files":["staged.txt"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/stage?workspace="+dir, body)
	rec := httptest.NewRecorder()
	handler.HandleGitStage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestParseGitStatus(t *testing.T) {
	input := `?? new.txt
 M modified.txt
A  added.txt
D  deleted.txt`

	statuses := parseGitStatus(input)
	if len(statuses) != 4 {
		t.Fatalf("expected 4 statuses, got %d", len(statuses))
	}

	expected := map[string]string{
		"new.txt":      "untracked",
		"modified.txt": "modified",
		"added.txt":    "added",
		"deleted.txt":  "deleted",
	}

	for _, s := range statuses {
		if exp, ok := expected[s.Path]; ok {
			if s.Status != exp {
				t.Errorf("file %s: expected status %q, got %q", s.Path, exp, s.Status)
			}
		}
	}
}

func TestParseGitLog(t *testing.T) {
	input := "abc123|John|2024-01-01|Initial commit\ndef456|Jane|2024-01-02|Add feature"

	commits := parseGitLog(input)
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	if commits[0].Hash != "abc123" || commits[0].Author != "John" {
		t.Errorf("unexpected first commit: %+v", commits[0])
	}
}

func TestParseGitBranches(t *testing.T) {
	input := "* main\n  feature/test\n  develop"

	branches := parseGitBranches(input)
	if len(branches) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(branches))
	}

	if !branches[0].Current || branches[0].Name != "main" {
		t.Errorf("expected main as current, got: %+v", branches[0])
	}
	if branches[1].Current {
		t.Errorf("expected feature/test not current")
	}
}
