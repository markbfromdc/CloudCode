# User Stories Specification

> **Version:** 1.0.0
> **Last Updated:** 2026-03-31
> **Status:** Approved

---

## 1. Personas

| Persona | Description |
|---------|-------------|
| **Developer** | End user who creates workspaces, edits code, runs commands, and commits changes |
| **Platform Admin** | Operator who deploys, configures, and monitors the CloudCode platform |

---

## 2. Workspace Management

### US-001: Create Workspace `[P0]`

**As a** Developer
**I want to** create a new workspace
**So that** I get an isolated development environment with my tools pre-installed

**Acceptance Criteria:**
- [ ] Clicking "Create Workspace" sends `POST /api/v1/workspaces`
- [ ] A Docker container is provisioned with Node 20, Python 3.11, Go 1.22, git, vim, tmux
- [ ] Container has memory limit (default 4GB), CPU limit, PID limit (512)
- [ ] Container has `no-new-privileges` security option
- [ ] Response returns `session_id`, `container_id`, and `status: "running"`
- [ ] Session is stored in the manager's session map with `CreatedAt` timestamp
- [ ] File tree loads from the workspace's `/workspace` directory

**Edge Cases:**
- Docker daemon unreachable: returns 500 with descriptive error
- Image not found: container create fails, error returned to user
- Container starts but immediately crashes: start failure triggers cleanup (container removed)
- Rate limit exceeded: returns 429 before reaching handler

**Workflow:**
```
Developer -> Create Workspace -> Container provisioned -> File tree loads -> Terminal connects
```

**Traceability:** API-002, ARCH-001, TEST-001

---

### US-002: Stop Workspace `[P0]`

**As a** Developer
**I want to** stop my workspace when I'm done
**So that** resources are freed

**Acceptance Criteria:**
- [ ] Sending `POST /api/v1/workspaces/stop?session_id=X` stops the container
- [ ] Container is gracefully stopped (10s timeout) then force-removed
- [ ] Session is removed from the manager's session map
- [ ] WebSocket connections for that session are terminated
- [ ] Health endpoint reflects decreased `active_workspaces` count

**Edge Cases:**
- Session not found: returns 500
- Container already stopped: remove still succeeds (force flag)
- Graceful stop times out: force remove proceeds

**Traceability:** API-003, TEST-002

---

### US-003: Session Timeout `[P1]`

**As a** Platform Admin
**I want** idle workspaces to be automatically stopped after a configurable timeout
**So that** resources aren't wasted on abandoned sessions

**Acceptance Criteria:**
- [ ] Cleanup loop runs at configurable interval (default: every 5 minutes)
- [ ] Sessions older than `CONTAINER_TIMEOUT_MIN` (default: 480 minutes) are stopped
- [ ] Cleanup calls `StopWorkspace()` for each expired session
- [ ] Cleanup loop stops cleanly during graceful shutdown

**Traceability:** ARCH-001, STATE-002

---

## 3. File Operations

### US-004: Browse File Tree `[P0]`

**As a** Developer
**I want to** see my workspace files in a tree view
**So that** I can navigate the project structure

**Acceptance Criteria:**
- [ ] File tree displays directories and files with correct hierarchy
- [ ] Directories are expandable/collapsible with chevron icons
- [ ] Files show color-coded icons by extension (blue for .ts, yellow for .js, etc.)
- [ ] Hidden files (starting with `.`) are excluded
- [ ] `node_modules`, `__pycache__`, `vendor` directories are excluded
- [ ] Tree depth limited to 3 levels
- [ ] Empty workspace shows "No files" message

**Edge Cases:**
- Workspace with 1000+ files: tree limited to 3 levels, browser remains responsive
- Symlinks: `filepath.Clean` resolves them safely
- Permission denied on a subdirectory: that subtree is skipped

**Traceability:** API-004, STATE-005

---

### US-005: Open and Edit Files `[P0]`

**As a** Developer
**I want to** click a file to open it in the editor
**So that** I can read and modify its content

**Acceptance Criteria:**
- [ ] Clicking a file in the explorer opens it in a new editor tab
- [ ] File content is loaded from the API (`readFile`)
- [ ] If the file is already open, the existing tab is activated (no duplicate)
- [ ] Editor uses Monaco with syntax highlighting based on file extension
- [ ] Tab shows a dot indicator when content is modified (dirty)
- [ ] Multiple files can be open simultaneously as tabs
- [ ] Ctrl+1-9 switches between tabs by index

**Edge Cases:**
- File > 10MB: API returns 413, tab opens with empty content
- API unavailable: tab opens with empty content (graceful degradation)
- Binary file: displayed as text (Monaco handles gracefully)
- File deleted externally while open: tab content preserved in memory

**Traceability:** API-005, STATE-005, STATE-006

---

### US-006: Save Files `[P0]`

**As a** Developer
**I want to** save my changes with Ctrl+S
**So that** modifications are persisted to the workspace filesystem

**Acceptance Criteria:**
- [ ] Pressing Ctrl+S triggers `writeFile` API call with current content
- [ ] On success, tab's `isDirty` flag is cleared
- [ ] "Saved" indicator appears briefly (1.5 seconds)
- [ ] Parent directories are created automatically if needed
- [ ] When no session is active, save only clears the dirty flag locally

**Edge Cases:**
- Save fails (API error): tab stays dirty, no error shown (silent failure)
- Rapid Ctrl+S: each save is independent (no debounce needed — saves are idempotent)
- Save while content is loading: current content is saved (may be empty)

**Traceability:** API-006, TEST-003

---

### US-007: Create, Delete, Rename Files `[P1]`

**As a** Developer
**I want to** create new files/folders, delete them, and rename them
**So that** I can manage my project structure

**Acceptance Criteria:**
- [ ] Create file: `POST /api/v1/files/create` with path and type
- [ ] Create directory: same endpoint with `type: "directory"`, creates nested dirs
- [ ] Delete: `DELETE /api/v1/files/delete?path=X`, works for files and directories
- [ ] Rename: `POST /api/v1/files/rename` with old_path and new_path
- [ ] Cannot delete workspace root (returns 403)
- [ ] Cannot create/delete/rename outside workspace (returns 403)

**Traceability:** API-007, API-008, API-009

---

## 4. Git Operations

### US-008: View Git Status `[P0]`

**As a** Developer
**I want to** see which files have changed in my workspace
**So that** I can track my modifications

**Acceptance Criteria:**
- [ ] Git panel shows current branch name (from `getGitBranches`)
- [ ] Changed files listed with status icons (modified=yellow, added=green, deleted=red, untracked=gray)
- [ ] File count shown next to "Changes" header
- [ ] Refresh button reloads status, branches, and log
- [ ] Loading spinner shows during refresh

**Edge Cases:**
- No git repo initialized: all lists empty, "No commits yet" shown
- Large number of changed files: all listed (no pagination)
- API unavailable: keeps previous state, no error shown

**Traceability:** API-010, API-012

---

### US-009: Commit Changes `[P0]`

**As a** Developer
**I want to** stage and commit my changes
**So that** I can save my work history

**Acceptance Criteria:**
- [ ] "Stage All" button calls `stageFiles([])` (stages all changes)
- [ ] Commit message input field with placeholder text
- [ ] Ctrl+Enter in message field triggers commit
- [ ] Commit button calls `createCommit(message)` then refreshes status
- [ ] Commit button disabled when message is empty
- [ ] After successful commit, message field is cleared
- [ ] Commits list updates to show new commit

**Edge Cases:**
- Empty message: commit button disabled, API rejects with 400
- Nothing to commit: git returns error, silently handled
- Commit message with special characters: handled safely

**Traceability:** API-013, API-014

---

## 5. Terminal

### US-010: Use Terminal `[P0]`

**As a** Developer
**I want to** use a terminal connected to my workspace container
**So that** I can run commands, build, and test

**Acceptance Criteria:**
- [ ] Terminal panel opens with xterm.js
- [ ] WebSocket connects to `/ws/terminal?session_id=X`
- [ ] Keystrokes are forwarded to container's bash stdin
- [ ] Container stdout is streamed back to the terminal
- [ ] Terminal auto-fits to available space
- [ ] Ctrl+` toggles terminal panel visibility

**Edge Cases:**
- No active session: terminal runs in local echo demo mode
- WebSocket disconnects: automatic reconnection with exponential backoff (max 5 attempts)
- Network timeout: pong timeout (40s) triggers disconnect
- Send buffer full (256 messages): excess messages dropped with warning log

**Traceability:** API-016, ARCH-004

---

## 6. Search

### US-011: Search Across Files `[P1]`

**As a** Developer
**I want to** search for text across all open files
**So that** I can find code quickly

**Acceptance Criteria:**
- [ ] Search input in sidebar search panel
- [ ] Results update as user types (reactive via `useMemo`)
- [ ] Results grouped by file with match count
- [ ] Each result shows line number and highlighted match
- [ ] Clicking a result switches to that file's tab
- [ ] Case sensitivity toggle works
- [ ] Regex toggle works (invalid regex handled gracefully)
- [ ] Replace button replaces all matches in the active file

**Edge Cases:**
- Empty query: no results shown
- Invalid regex: returns empty results (no error)
- Zero-length regex match: prevented from infinite loop (lastIndex incremented)
- Large file with many matches: limited to 50 matches per file in UI

**Traceability:** STATE-005

---

## 7. Command Palette & Keyboard Shortcuts

### US-012: Quick File Open `[P1]`

**As a** Developer
**I want to** quickly open files by name using Ctrl+P
**So that** I can navigate without browsing the file tree

**Acceptance Criteria:**
- [ ] Ctrl+P opens command palette in file search mode
- [ ] Files filtered by name and path as user types
- [ ] Maximum 20 results shown
- [ ] Arrow keys navigate, Enter opens selected file
- [ ] Escape closes palette
- [ ] File content loaded from API on open

**Traceability:** STATE-005

---

### US-013: Command Palette `[P1]`

**As a** Developer
**I want to** access IDE commands via Ctrl+Shift+P
**So that** I can toggle panels and navigate without mouse

**Acceptance Criteria:**
- [ ] Typing `>` prefix switches to command mode
- [ ] Available commands: Toggle Terminal, Toggle Sidebar, Show Explorer, Show Search, Show Source Control, Close Active Editor, Open Terminal, Show Problems
- [ ] Each command shows its keyboard shortcut
- [ ] Commands are filtered by typed text
- [ ] Selected command executes on Enter

**Keyboard shortcuts registered:**

| Shortcut | Action |
|----------|--------|
| `Ctrl+P` | Open command palette (file mode) |
| `Ctrl+Shift+P` | Open command palette (command mode) |
| `Ctrl+B` | Toggle sidebar |
| `` Ctrl+` `` | Toggle terminal |
| `Ctrl+W` | Close active tab |
| `Ctrl+S` | Save active file |
| `Ctrl+Shift+E` | Show Explorer |
| `Ctrl+Shift+F` | Show Search |
| `Ctrl+Shift+G` | Show Source Control |
| `Ctrl+1` through `Ctrl+9` | Switch to tab by index |

**Traceability:** STATE-005

---

## 8. Platform Administration

### US-014: Health Monitoring `[P0]`

**As a** Platform Admin
**I want to** check server health via `/health`
**So that** I can monitor uptime and resource usage

**Acceptance Criteria:**
- [ ] `/health` is accessible without authentication
- [ ] Returns active session and workspace counts
- [ ] Returns UTC timestamp
- [ ] Responds within 100ms under normal load

**Traceability:** API-001

---

### US-015: Rate Limiting `[P1]`

**As a** Platform Admin
**I want** API requests to be rate-limited per IP
**So that** the platform is protected from abuse

**Acceptance Criteria:**
- [ ] Token-bucket rate limiter with configurable RPS and burst
- [ ] Per-IP tracking with automatic cleanup of stale entries
- [ ] Returns 429 with appropriate error when exceeded
- [ ] `X-RateLimit-Remaining` header in all responses
- [ ] IP extracted from `X-Forwarded-For`, `X-Real-IP`, or `RemoteAddr`

**Traceability:** API-RATE-001, SEC-001

---

### US-016: Graceful Shutdown `[P0]`

**As a** Platform Admin
**I want** the server to shut down gracefully on SIGTERM
**So that** active connections are drained and containers are cleaned up

**Acceptance Criteria:**
- [ ] HTTP server stops accepting new connections
- [ ] Existing connections drain within `SHUTDOWN_TIMEOUT_SEC` (default 30s)
- [ ] WebSocket hub stops and disconnects all clients
- [ ] Container manager stops all running workspace containers
- [ ] Rate limiter cleanup goroutine is stopped
- [ ] Process exits cleanly with code 0

**Traceability:** ARCH-004

---

## Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-03-31 | CloudCode Team | Initial specification — 16 user stories |
