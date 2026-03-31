# Database & State Schema Specification

> **Version:** 1.0.0
> **Last Updated:** 2026-03-31
> **Status:** Approved

---

## 1. Overview `[STATE-001]`

CloudCode currently uses **in-memory state** for all runtime data. There is no persistent database. State is partitioned between the Go backend (server-side) and the React frontend (client-side). This document defines both.

---

## 2. Backend State — Container Manager `[STATE-002]`

### 2.1 Entity: WorkspaceSession

The primary backend entity. Stored in `Manager.sessions` (type: `map[string]*WorkspaceSession`, protected by `sync.RWMutex`).

**Source:** `internal/container/manager.go`

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `SessionID` | `string` | UUID v4, primary key | `"550e8400-e29b-..."` |
| `ContainerID` | `string` | Docker container SHA | `"a1b2c3d4e5f6..."` |
| `UserID` | `string` | From JWT `sub` claim | `"user-42"` |
| `Status` | `string` | Container state | `"running"` |
| `CreatedAt` | `time.Time` | Session creation timestamp | `2026-03-31T12:00:00Z` |

**Relationships:**

```
WorkspaceSession 1 --- 1 Docker Container
WorkspaceSession N --- 1 User (via UserID)
WorkspaceSession 1 --- N WebSocket Clients (via Hub)
```

**Lifecycle:**
1. Created by `Manager.CreateWorkspace()` — sets `Status = "running"`, `CreatedAt = time.Now()`
2. Queried by `Manager.GetSession()` — read-only lookup
3. Deleted by `Manager.StopWorkspace()` — removes from map, stops container
4. Expired by `Manager.StartCleanupLoop()` — removes sessions older than `maxAge`
5. Bulk cleanup by `Manager.Shutdown()` — stops all sessions on server shutdown

**Indexing strategy:**
- Primary index: `SessionID` (map key) — O(1) lookup
- No secondary indexes (UserID lookup requires full scan)

**Concurrency:** All access protected by `sync.RWMutex`. `GetSession()` and `ActiveWorkspaces()` use `RLock`. `CreateWorkspace()`, `StopWorkspace()`, and `Shutdown()` use `Lock`.

### 2.2 Entity: ExecSession

Transient entity representing an active terminal attachment. Not stored persistently.

**Source:** `internal/container/manager.go`

| Field | Type | Description |
|-------|------|-------------|
| `ContainerID` | `string` | Docker container ID |
| `Stdin` | `io.WriteCloser` | Write to container stdin |
| `Stdout` | `io.Reader` | Read from container stdout |

**Lifecycle:** Created by `Manager.AttachToContainer()`, used by WebSocket handler, closed when client disconnects.

---

## 3. Backend State — WebSocket Hub `[STATE-003]`

### 3.1 Entity: Client Registry

The Hub maintains a map of active WebSocket clients.

**Source:** `internal/websocket/hub.go`

**Storage:** `Hub.clients` (type: `map[string]*Client`, protected by `sync.RWMutex`)

| Field | Type | Description |
|-------|------|-------------|
| Key | `string` | SessionID |
| Value | `*Client` | WebSocket client instance |

**Client fields relevant to state:**

| Field | Type | Description |
|-------|------|-------------|
| `SessionID` | `string` | Links to WorkspaceSession |
| `UserID` | `string` | Authenticated user |
| `ContainerID` | `string` | Docker container backing this session |
| `send` | `chan []byte` | Buffered channel (capacity: 256) |

**Operations:**
- `Register(client)` — adds to map
- `Unregister(client)` — removes from map, closes send channel
- `GetClient(sessionID)` — read-only lookup
- `ActiveSessions()` — returns count
- `Stop()` — closes all clients, stops run loop

---

## 4. Backend State — Rate Limiter `[STATE-004]`

**Source:** `internal/middleware/ratelimit.go`

**Storage:** `RateLimiter.visitors` (type: `sync.Map`)

| Key | Value | Description |
|-----|-------|-------------|
| IP string | `*visitor{limiter, lastSeen}` | Per-IP rate limiter state |

**Cleanup:** Background goroutine runs every 3 minutes, evicts entries not seen in the last 3 minutes.

---

## 5. Frontend State — WorkspaceContext `[STATE-005]`

### 5.1 State Shape

The React frontend manages all UI state through a single `useReducer` in `WorkspaceContext`.

**Source:** `frontend/src/context/WorkspaceContext.tsx`

```typescript
interface WorkspaceState {
  sessionId: string | null;        // Active workspace session UUID
  files: FileNode[];                // File tree hierarchy
  openTabs: EditorTab[];            // Currently open editor tabs
  activeTabId: string | null;       // ID of the focused tab
  activeActivity: ActivityBarItem;  // Current sidebar panel
  activePanel: PanelTab;            // Current bottom panel
  isPanelOpen: boolean;             // Bottom panel visibility
  isSidebarOpen: boolean;           // Sidebar visibility
  isConnected: boolean;             // WebSocket connection status
  isLoading: boolean;               // Global loading indicator
  error: string | null;             // Global error message
}
```

**Initial state:**

| Field | Default |
|-------|---------|
| `sessionId` | `null` |
| `files` | `[]` |
| `openTabs` | `[]` |
| `activeTabId` | `null` |
| `activeActivity` | `'explorer'` |
| `activePanel` | `'terminal'` |
| `isPanelOpen` | `true` |
| `isSidebarOpen` | `true` |
| `isConnected` | `false` |
| `isLoading` | `false` |
| `error` | `null` |

### 5.2 Actions (Reducer)

| Action | Payload | Effect |
|--------|---------|--------|
| `SET_SESSION` | `{ sessionId: string }` | Sets active session |
| `SET_FILES` | `{ files: FileNode[] }` | Replaces file tree |
| `OPEN_FILE` | `{ tab: EditorTab }` | Adds tab (or activates existing) |
| `CLOSE_TAB` | `{ tabId: string }` | Removes tab; selects last remaining |
| `SET_ACTIVE_TAB` | `{ tabId: string }` | Changes active tab |
| `UPDATE_TAB_CONTENT` | `{ tabId, content }` | Updates content + sets `isDirty: true` |
| `SET_TAB_CONTENT` | `{ tabId, content }` | Updates content without marking dirty |
| `MARK_TAB_SAVED` | `{ tabId }` | Sets `isDirty: false` |
| `SET_ACTIVITY` | `{ activity }` | Changes sidebar panel; toggles if same |
| `SET_PANEL` | `{ panel }` | Opens bottom panel to specified tab |
| `TOGGLE_PANEL` | — | Toggles bottom panel visibility |
| `TOGGLE_SIDEBAR` | — | Toggles sidebar visibility |
| `SET_CONNECTED` | `{ connected: boolean }` | Updates WebSocket status |
| `TOGGLE_FILE_EXPAND` | `{ path: string }` | Toggles directory expansion (recursive) |
| `SET_LOADING` | `{ isLoading: boolean }` | Sets global loading state |
| `SET_ERROR` | `{ error: string | null }` | Sets or clears global error |

### 5.3 Key Business Logic in Reducer

**`OPEN_FILE` deduplication:** If a tab with the same `id` already exists, it is activated without creating a duplicate.

**`CLOSE_TAB` fallback:** When closing the active tab, the last remaining tab becomes active. If no tabs remain, `activeTabId` becomes `null`.

**`SET_ACTIVITY` toggle:** Clicking the already-active activity icon toggles the sidebar closed. Clicking a different icon opens the sidebar with that panel.

**`TOGGLE_FILE_EXPAND` recursion:** The `toggleExpand` helper recursively searches the file tree by path and toggles `isExpanded` on the matching node.

---

## 6. Shared Type Definitions `[STATE-006]`

**Source:** `frontend/src/types/index.ts`

### 6.1 FileNode

```typescript
interface FileNode {
  name: string;         // Display name ("main.ts")
  path: string;         // Absolute path ("/workspace/src/main.ts")
  type: 'file' | 'directory';
  children?: FileNode[];  // Present only for directories
  isExpanded?: boolean;   // UI expansion state
}
```

### 6.2 EditorTab

```typescript
interface EditorTab {
  id: string;           // Unique identifier (typically the file path)
  path: string;         // File path for API calls
  name: string;         // Display name for tab label
  language: string;     // Monaco Editor language ID
  content: string;      // File content (loaded from API)
  isDirty: boolean;     // True when content differs from saved version
}
```

### 6.3 WorkspaceSession (API response)

```typescript
interface WorkspaceSession {
  session_id: string;
  container_id: string;
  status: string;
}
```

### 6.4 HealthStatus (API response)

```typescript
interface HealthStatus {
  status: string;
  active_sessions: number;
  active_workspaces: number;
  timestamp: string;
}
```

### 6.5 WebSocket Messages

```typescript
type WSMessageType = 'input' | 'output' | 'resize' | 'heartbeat';

interface WSMessage {
  type: WSMessageType;
  data?: string;
  cols?: number;
  rows?: number;
}
```

### 6.6 UI Type Literals

```typescript
type ActivityBarItem = 'explorer' | 'search' | 'git' | 'extensions' | 'settings';
type PanelTab = 'terminal' | 'output' | 'problems';
```

---

## 7. Client-Side Storage `[STATE-007]`

| Key | Storage | Type | Description |
|-----|---------|------|-------------|
| `cloudcode_token` | `localStorage` | string | JWT bearer token |

No other client-side persistence is used. All workspace state is ephemeral and lost on page refresh.

---

## 8. Future Database Schema (Planned) `[STATE-008]`

When persistent storage is added, the following schema is recommended:

### 8.1 Users Table

```sql
CREATE TABLE users (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email       VARCHAR(255) UNIQUE NOT NULL,
  name        VARCHAR(255),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users(email);
```

### 8.2 Workspaces Table

```sql
CREATE TABLE workspaces (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  container_id  VARCHAR(128),
  status        VARCHAR(32) NOT NULL DEFAULT 'running',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  stopped_at    TIMESTAMPTZ,
  last_active   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_workspaces_user ON workspaces(user_id);
CREATE INDEX idx_workspaces_status ON workspaces(status);
CREATE INDEX idx_workspaces_created ON workspaces(created_at);
```

### 8.3 Audit Log Table

```sql
CREATE TABLE audit_log (
  id          BIGSERIAL PRIMARY KEY,
  user_id     UUID REFERENCES users(id),
  action      VARCHAR(64) NOT NULL,
  resource    VARCHAR(255),
  details     JSONB,
  ip_address  INET,
  request_id  UUID,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_created ON audit_log(created_at);
```

### 8.4 Entity Relationship Diagram

```
+--------+       +-------------+       +-----------+
| users  | 1---N | workspaces  |       | audit_log |
|--------|       |-------------|       |-----------|
| id     |       | id          |       | id        |
| email  |       | user_id  FK |       | user_id   |
| name   |       | container_id|       | action    |
+--------+       | status      |       | resource  |
                 | created_at  |       | details   |
                 | stopped_at  |       | ip_address|
                 +-------------+       | request_id|
                                       +-----------+
```

---

## Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-03-31 | CloudCode Team | Initial specification |
