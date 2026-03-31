# Testing Specification

> **Version:** 1.0.0
> **Last Updated:** 2026-03-31
> **Status:** Approved

---

## 1. Overview `[TEST-OVERVIEW]`

| Metric | Value |
|--------|-------|
| Total tests | 195 |
| Backend tests (Go) | 109 |
| Frontend tests (Vitest) | 86 |
| Test files | 20 |
| All passing | Yes |
| Race detector | Enabled for all Go tests |

---

## 2. Test Strategy `[TEST-STRATEGY]`

### 2.1 Test Pyramid

```
          /   E2E   \           Planned (not yet implemented)
         /  (manual)  \
        /--------------\
       / Integration    \       Backend: middleware chain tests
      /   Tests          \      Frontend: component render tests
     /--------------------\
    /     Unit Tests       \    109 Go + 86 Vitest = 195 tests
   /________________________\
```

### 2.2 Principles

1. **Race detection:** All Go tests run with `-race` flag
2. **Isolation:** Tests use `t.TempDir()` for filesystem operations, mock interfaces for Docker
3. **No external dependencies:** Tests never require Docker daemon, network, or database
4. **Deterministic:** No flaky tests, no time-dependent assertions (except explicit timeouts)
5. **Fast:** Full suite runs in <5 seconds

---

## 3. Backend Unit Tests `[TEST-001]`

### 3.1 Package: `internal/api` — 46 tests

**File tree tests** (`filetree_test.go`):

| Test | Description | Requirement |
|------|-------------|-------------|
| TestHandleListFiles | Lists files in temp directory | API-004 |
| TestHandleListFilesMethodNotAllowed | Rejects non-GET | API-004 |
| TestHandleReadFile | Reads file content | API-005 |
| TestHandleReadFileNotFound | Returns 404 for missing file | API-005 |
| TestHandleReadFileAccessDenied | Blocks path traversal | API-005 |
| TestHandleWriteFile | Writes content to file | API-006 |
| TestHandleWriteFileMissingPath | Rejects missing path param | API-006 |
| TestBuildFileTreeSkipsHidden | Excludes dotfiles | API-004 |

**File operations tests** (`files_test.go`):

| Test | Description | Requirement |
|------|-------------|-------------|
| TestHandleCreateFile | Creates new file | API-007 |
| TestHandleCreateDirectory | Creates nested directories | API-007 |
| TestHandleDeleteFile | Deletes file | API-008 |
| TestHandleDeleteFileAccessDenied | Blocks path traversal | API-008 |
| TestHandleDeleteWorkspaceRoot | Prevents root deletion | API-008 |
| TestHandleRenameFile | Renames file | API-009 |
| TestHandleRenameFileAccessDenied | Blocks path traversal on rename | API-009 |

**Git tests** (`git_test.go`):

| Test | Description | Requirement |
|------|-------------|-------------|
| TestHandleGitStatus | Returns working tree status | API-010 |
| TestHandleGitLog | Returns commit history | API-011 |
| TestHandleGitBranches | Returns branch list | API-012 |
| TestHandleGitCommit | Creates commit with staged files | API-013 |
| TestHandleGitCommitEmptyMessage | Rejects empty message | API-013 |
| TestHandleGitInit | Initializes git repo | API-015 |
| TestHandleGitStage | Stages files | API-014 |
| TestParseGitStatus | Parses porcelain output | API-010 |
| TestParseGitLog | Parses log format | API-011 |
| TestParseGitBranches | Parses branch list | API-012 |

**Edge case tests** (`edge_test.go`):

| Test | Description | Requirement |
|------|-------------|-------------|
| TestHandleReadFileTooLarge | Rejects files > 10MB | API-005 |
| TestHandleWriteFileEmptyContent | Allows empty content | API-006 |
| TestHandleListFilesEmpty | Handles empty directory | API-004 |
| TestHandleGit*MethodNotAllowed (7) | Rejects wrong HTTP methods | API-010..015 |
| TestHandleGitStageAll | Stages all with empty files array | API-014 |
| TestHandleGitCommitInvalidJSON | Rejects malformed body | API-013 |
| TestResolveWorkDirSecurity | Validates path traversal protection | API-010 |
| TestHandleCreate/Delete/Rename*MethodNotAllowed (3) | HTTP method checks | API-007..009 |
| TestHandleCreateFileEmptyPath | Rejects empty path | API-007 |
| TestHandleCreateFileAccessDenied | Blocks path traversal | API-007 |
| TestHandleDeleteFileMissingPath | Rejects missing path | API-008 |
| TestHandleRenameFileInvalidJSON | Rejects malformed body | API-009 |

### 3.2 Package: `internal/container` — 18 tests

| Test | Description | Mock |
|------|-------------|------|
| TestGetSession | Session lookup (found and not found) | — |
| TestActiveWorkspaces | Returns correct count | — |
| TestCreateWorkspaceSuccess | Container created with correct config | DockerClient |
| TestCreateWorkspaceEnvVars | Verifies SESSION_ID and USER_ID env vars | DockerClient |
| TestCreateWorkspaceCreateFails | Handles docker create error | DockerClient |
| TestCreateWorkspaceStartFailsCleansUp | Removes container on start failure | DockerClient |
| TestStopWorkspaceSuccess | Stops and removes container | DockerClient |
| TestStopWorkspaceNotFound | Returns error for missing session | — |
| TestStopWorkspaceRemoveFails | Handles remove failure | DockerClient |
| TestStopWorkspaceGracefulStopFails | Proceeds to remove even if stop fails | DockerClient |
| TestAttachToContainerSuccess | Creates exec and attaches | DockerClient |
| TestAttachToContainerNotFound | Returns error for missing session | — |
| TestAttachToContainerExecCreateFails | Handles exec create error | DockerClient |
| TestAttachToContainerExecAttachFails | Handles exec attach error | DockerClient |
| TestConcurrentSessionAccess | Race condition test (50 goroutines) | — |
| TestContainerExecConfig | Verifies exec config builder | — |
| TestInt64Ptr | Pointer helper function | — |
| TestWorkspaceSessionFields | Struct field assignment | — |

**Mock:** `mockDockerClient` implements the `DockerClient` interface (6 methods) with configurable function callbacks.

### 3.3 Package: `internal/config` — 6 tests

| Test | Description |
|------|-------------|
| TestLoadRequiresJWTSecret | Fails without JWT_SECRET |
| TestLoadWithValidConfig | Loads defaults correctly |
| TestLoadCustomPorts | Overrides ports from env |
| TestHTTPAddr | Formats address string |
| TestGRPCAddr | Formats gRPC address |
| TestTLSValidation | Validates cert/key file existence |

### 3.4 Package: `internal/middleware` — 10 + new tests

**Auth tests** (`auth_test.go`):

| Test | Description |
|------|-------------|
| TestAuthMiddlewareValidToken | Passes through with valid JWT |
| TestAuthMiddlewareMissingHeader | Returns 401 |
| TestAuthMiddlewareExpiredToken | Returns 401 |
| TestAuthMiddlewareInvalidSignature | Returns 401 |
| TestAuthMiddlewareInvalidFormat | Returns 401 |

**CORS tests** (`cors_test.go`):

| Test | Description |
|------|-------------|
| TestCORSAllowedOrigin | Sets CORS headers |
| TestCORSBlockedOrigin | No CORS headers |
| TestCORSPreflight | Returns 204 for OPTIONS |

**Logging tests** (`logging_test.go`):

| Test | Description |
|------|-------------|
| TestRequestLogger | Logs method, path, status |
| TestRequestLoggerCapturesStatusCode | Captures 404 status |

**Rate limit tests** (`ratelimit_test.go`):

| Test | Description |
|------|-------------|
| TestRateLimiterAllowsBelowLimit | Requests pass under limit |
| TestRateLimiterRejectsOverLimit | Returns 429 when exceeded |
| TestRateLimiterSetsHeader | X-RateLimit-Remaining header present |
| TestRateLimiterPerIP | Different IPs have independent limits |
| TestRateLimiterIPExtraction | Extracts IP from headers and RemoteAddr |

**Request ID tests** (`requestid_test.go`):

| Test | Description |
|------|-------------|
| TestRequestIDHeaderSet | X-Request-ID header in response |
| TestRequestIDIsUUID | Format validation |
| TestRequestIDUnique | Different per request |
| TestGetRequestIDFromContext | Context value retrieval |

### 3.5 Package: `internal/websocket` — 13 tests

| Test | Description |
|------|-------------|
| TestNewHub | Hub initialization |
| TestHubRegisterUnregister | Client registration lifecycle |
| TestHubGetClientNotFound | Missing client lookup |
| TestNewClient | Client construction |
| TestClientSendBufferFull | Dropped messages on full buffer |
| TestClientSendQueuesMessage | Message queued in send channel |
| TestNewHandler | Handler initialization |
| TestServeHTTPMissingSessionID | Returns 401 |
| TestServeHTTPMissingUserContext | Returns 401 |
| TestServeHTTPInvalidUserContext | Returns 500 |
| TestServeHTTPNoWebSocketUpgrade | Handles non-WS request |
| TestContextKeyUserID | Context key value |
| TestUpgraderOriginCheck | Origin validation (allowed + blocked) |

### 3.6 Package: `internal/logging` — 5 tests

| Test | Description |
|------|-------------|
| TestLoggerLevels | Level filtering (debug < info < warn < error) |
| TestLoggerWithField | Field injection |
| TestLoggerNilWriter | Nil writer safety |
| TestLogLevelString | Level string representation |
| TestDefaultLogger | Default logger construction |

---

## 4. Frontend Unit Tests `[TEST-002]`

### 4.1 WorkspaceContext Reducer — 20 tests

**Source:** `frontend/src/context/WorkspaceContext.test.tsx`

| Test | Action Tested | Assertion |
|------|---------------|-----------|
| provides initial state | — | All defaults correct |
| throws outside provider | `useWorkspace()` | Error thrown |
| SET_SESSION | sessionId | Updated |
| SET_FILES | files | Replaced |
| OPEN_FILE (new) | tab | Added + active |
| OPEN_FILE (existing) | tab | Deduplicated |
| CLOSE_TAB (active) | tabId | Removed, last tab active |
| CLOSE_TAB (last) | tabId | Removed, activeTabId null |
| CLOSE_TAB (non-active) | tabId | Removed, active unchanged |
| SET_ACTIVE_TAB | tabId | Changed |
| UPDATE_TAB_CONTENT | tabId + content | Content updated, isDirty true |
| MARK_TAB_SAVED | tabId | isDirty false |
| SET_ACTIVITY (different) | activity | Changed, sidebar open |
| SET_ACTIVITY (same) | activity | Sidebar toggled |
| SET_PANEL | panel | Set, panel opens |
| TOGGLE_PANEL | — | Flipped |
| TOGGLE_SIDEBAR | — | Flipped |
| SET_CONNECTED | connected | Updated |
| TOGGLE_FILE_EXPAND (root) | path | isExpanded toggled |
| TOGGLE_FILE_EXPAND (nested) | path | Nested node toggled |

### 4.2 API Service — 10 tests

**Source:** `frontend/src/services/api.test.ts`

| Test | Function | Assertion |
|------|----------|-----------|
| sends POST to /workspaces | createWorkspace | Correct endpoint + method |
| sends POST with session_id | stopWorkspace | URL encoding |
| encodes special characters | stopWorkspace | %20 encoding |
| fetches /health | getHealth | Correct endpoint |
| fetches file list | listFiles | Path encoding |
| defaults to root path | listFiles | Default `/` |
| fetches file content as text | readFile | Returns text, auth headers sent |
| throws on read error | readFile | Error thrown with status |
| sends PUT with content | writeFile | Correct body format |
| throws on non-ok response | createWorkspace | Error with status + body |

### 4.3 Git Service — 14 tests

**Source:** `frontend/src/services/git.test.ts`

| Test | Function | Assertion |
|------|----------|-----------|
| fetches git status | getGitStatus | Correct endpoint, auth headers |
| passes workspace parameter | getGitStatus | URL encoding |
| returns empty on error | getGitStatus | Graceful fallback |
| fetches commit log | getGitLog | Returns array |
| returns empty on error | getGitLog | Graceful fallback |
| fetches branches | getGitBranches | Returns array |
| returns empty on error | getGitBranches | Graceful fallback |
| sends POST with files | stageFiles | Correct body |
| passes workspace parameter | stageFiles | URL encoding |
| sends POST with message | createCommit | Correct body |
| includes optional files | createCommit | Body includes files |
| throws on error | createCommit | Error thrown |
| sends POST to init | initRepo | Correct endpoint |
| passes workspace parameter | initRepo | URL encoding |

### 4.4 WebSocket Service — 7 tests

**Source:** `frontend/src/services/websocket.test.ts`

| Test | Description |
|------|-------------|
| creates with correct session id | Constructor |
| constructs correct WebSocket URL | URL includes session_id |
| calls onStatus(true) on open | Connection callback |
| disconnect prevents reconnection | Cleanup |
| disconnect safe without connect | No-throw |
| send does nothing when not connected | No-throw |
| resize does nothing when not connected | No-throw |

### 4.5 useFileLanguage Hook — 35 tests

**Source:** `frontend/src/hooks/useFileLanguage.test.ts`

Tests `getLanguageForFile()` for all supported extensions: `.ts`, `.tsx`, `.js`, `.jsx`, `.py`, `.go`, `.rs`, `.json`, `.html`, `.css`, `.scss`, `.md`, `.yml`, `.yaml`, `.sh`, `.sql`, `.c`, `.cpp`, `.h`, `.java`, `.rb`, `.php`, `.graphql`, `.xml`, `.toml`, `.txt`, `.env`, `.gitignore`, plus special cases for `Dockerfile`, `Makefile`, case insensitivity, and multi-dot filenames.

---

## 5. Integration Test Requirements `[TEST-003]`

### 5.1 Planned Backend Integration Tests

| Test | Description | Components |
|------|-------------|------------|
| Full auth flow | JWT creation → request → auth middleware → handler | middleware + api |
| Rate limit enforcement | Burst requests → 429 after limit | middleware + server |
| CORS preflight + actual | OPTIONS then GET with Origin | middleware |
| File CRUD lifecycle | Create → write → read → rename → delete | api/filetree + api/files |
| Git workflow | Init → create file → stage → commit → status → log | api/git |
| Health check unauthenticated | GET /health without JWT → 200 | server |

### 5.2 Planned Frontend Integration Tests

| Test | Description | Components |
|------|-------------|------------|
| File open flow | Click file → tab opens → content loads | FileExplorer + CodeEditor + API |
| Save flow | Edit → Ctrl+S → saved indicator | CodeEditor + API |
| Git commit flow | Type message → commit → refresh | GitPanel + git.ts |
| Search flow | Type query → results shown → click result | SearchPanel + WorkspaceContext |
| Keyboard shortcuts | Ctrl+P → palette → select file → opens | useKeyboardShortcuts + CommandPalette |

---

## 6. Performance Benchmarks `[TEST-004]`

### 6.1 Backend Targets

| Metric | Target | Measured |
|--------|--------|----------|
| Health check latency (p99) | < 10ms | ~1ms |
| File read latency (1KB file) | < 50ms | ~5ms |
| File list latency (100 files) | < 100ms | ~20ms |
| WebSocket message latency | < 5ms | ~1ms |
| Container create latency | < 10s | Docker-dependent |
| Test suite execution time | < 10s | ~5s |

### 6.2 Frontend Targets

| Metric | Target | Measured |
|--------|--------|----------|
| Production bundle size | < 700KB | 603KB |
| Test suite execution time | < 5s | ~2s |
| TypeScript compilation | < 10s | ~3s |
| First contentful paint | < 2s | Vite-dependent |
| Monaco editor load | < 3s | Network-dependent |

---

## 7. Security Testing Protocols `[TEST-005]`

### 7.1 Authentication Tests

| Test | Description | Status |
|------|-------------|--------|
| Missing auth header | Returns 401 | Implemented |
| Expired token | Returns 401 | Implemented |
| Invalid signature | Returns 401 | Implemented |
| Malformed token format | Returns 401 | Implemented |
| Token with wrong algorithm | Should reject | Tested via signature check |

### 7.2 Authorization Tests

| Test | Description | Status |
|------|-------------|--------|
| Path traversal (file read) | Blocked by path validation | Implemented |
| Path traversal (file write) | Blocked by path validation | Implemented |
| Path traversal (file create) | Blocked by path validation | Implemented |
| Path traversal (file delete) | Blocked by path validation | Implemented |
| Path traversal (file rename) | Blocked by path validation | Implemented |
| Path traversal (git operations) | Blocked by workspace validation | Implemented |
| Workspace root deletion | Returns 403 | Implemented |

### 7.3 Rate Limiting Tests

| Test | Description | Status |
|------|-------------|--------|
| Below limit passes | Requests succeed | Implemented |
| Over limit rejected | Returns 429 | Implemented |
| Per-IP isolation | Different IPs independent | Implemented |
| Header extraction | X-Forwarded-For respected | Implemented |

### 7.4 Input Validation Tests

| Test | Description | Status |
|------|-------------|--------|
| File > 10MB rejected | Returns 413 | Implemented |
| Invalid JSON body | Returns 400 | Implemented |
| Empty required params | Returns 400 | Implemented |
| Wrong HTTP method | Returns 405 | Implemented (all endpoints) |

---

## 8. Test Execution `[TEST-006]`

### 8.1 Local Execution

```bash
# Backend (109 tests, ~5s)
make test
# or: go test -v -race -count=1 ./internal/...

# Frontend (86 tests, ~2s)
make test-frontend
# or: cd frontend && npm test

# All tests
make test-all

# Coverage report
make coverage
```

### 8.2 CI Execution

Tests run automatically on push and PR via GitHub Actions (`.github/workflows/ci.yml`):
- Backend: `go build` → `go vet` → `go test -v -race`
- Frontend: `npm ci` → `tsc --noEmit` → `npm test` → `vite build`

---

## Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-03-31 | CloudCode Team | Initial specification — 195 tests documented |
