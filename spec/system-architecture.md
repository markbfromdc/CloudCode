# System Architecture Specification

> **Version:** 1.0.0
> **Last Updated:** 2026-03-31
> **Status:** Approved

---

## 1. System Overview `[ARCH-001]`

CloudCode is a cloud-based IDE that provisions isolated Docker containers per workspace. Users access a VS Code-style editor, terminal, git, and file management through a browser. The system consists of a Go HTTP/WebSocket backend, a React single-page application frontend, and Docker workspace containers.

---

## 2. High-Level Architecture `[ARCH-002]`

```
+-------------------+          +----------------------------------+
|                   |   HTTP   |         Go Backend Server        |
|   Browser (SPA)   +--------->|                                  |
|                   |          |  +----------------------------+  |
|  React 19         |   REST   |  |     Middleware Chain        |  |
|  Monaco Editor    +--------->|  |  Logger -> RateLimit ->     |  |
|  xterm.js         |          |  |  RequestID -> CORS -> Auth  |  |
|  Tailwind CSS     |          |  +----------------------------+  |
|                   |          |                                  |
|                   |   WS     |  +----------------------------+  |
|                   +--------->|  |     WebSocket Hub           |  |
|                   |          |  |  Client registry, ping/pong |  |
|                   |          |  +----------------------------+  |
+-------------------+          |                                  |
                               |  +----------------------------+  |
                               |  |   Container Manager         |  |
                               |  |   Docker SDK, session mgmt  |  |
                               |  +-------------+--------------+  |
                               +----------------|------------------+
                                                |
                                                | Docker API
                                                v
                               +----------------------------------+
                               |     Docker Engine                |
                               |                                  |
                               |  +----------+  +----------+     |
                               |  |Workspace |  |Workspace |     |
                               |  |Container |  |Container |     |
                               |  |  (user1) |  |  (user2) |     |
                               |  |          |  |          |     |
                               |  | Node 20  |  | Node 20  |     |
                               |  | Python   |  | Python   |     |
                               |  | Go 1.22  |  | Go 1.22  |     |
                               |  | bash/vim |  | bash/vim |     |
                               |  +----------+  +----------+     |
                               +----------------------------------+
```

---

## 3. Component Diagram `[ARCH-003]`

### 3.1 Backend Components

```
cmd/server/main.go
  |
  +-- internal/config/        Config loading (env vars)
  |
  +-- internal/middleware/     HTTP middleware
  |     +-- auth.go           JWT validation (HMAC-SHA256)
  |     +-- cors.go           CORS header management
  |     +-- logging.go        Request/response logging
  |     +-- ratelimit.go      Token-bucket per-IP rate limiter
  |     +-- requestid.go      UUID correlation ID injection
  |
  +-- internal/api/           HTTP API handlers
  |     +-- filetree.go       File tree listing, read, write
  |     +-- files.go          File/directory create, delete, rename
  |     +-- git.go            Git status, log, branches, commit, stage, init
  |     +-- errors.go         Structured error response helpers
  |
  +-- internal/container/     Docker container lifecycle
  |     +-- manager.go        Create, attach, stop, cleanup, shutdown
  |
  +-- internal/websocket/     Real-time terminal communication
  |     +-- hub.go            Client registry, broadcast, stop
  |     +-- client.go         Read/write pumps, heartbeat, send buffer
  |     +-- handler.go        HTTP upgrade, exec attach, stream output
  |
  +-- internal/logging/       Structured leveled logger
        +-- logger.go         Info/Warn/Error/Debug with field support
```

### 3.2 Frontend Components

```
frontend/src/
  |
  +-- main.tsx                  React entry point
  +-- App.tsx                   Root: ErrorBoundary > WorkspaceProvider > IDELayout
  |
  +-- context/
  |     +-- WorkspaceContext.tsx  React Context + useReducer (16 action types)
  |
  +-- components/
  |     +-- layout/
  |     |     +-- IDELayout.tsx     Main layout orchestrator
  |     |     +-- TitleBar.tsx      Window title bar
  |     |     +-- ActivityBar.tsx   Left icon sidebar
  |     |     +-- BottomPanel.tsx   Terminal/output/problems tabs
  |     |     +-- StatusBar.tsx     Bottom status information
  |     |
  |     +-- editor/
  |     |     +-- CodeEditor.tsx    Monaco editor with Ctrl+S save
  |     |     +-- EditorTabs.tsx    Multi-tab bar with dirty indicators
  |     |
  |     +-- sidebar/
  |     |     +-- Sidebar.tsx          Panel router
  |     |     +-- FileExplorer.tsx     Recursive file tree with API loading
  |     |     +-- SearchPanel.tsx      Client-side search across open tabs
  |     |     +-- GitPanel.tsx         Git status/commit with API integration
  |     |     +-- ExtensionsPanel.tsx  Extension marketplace UI
  |     |     +-- SettingsPanel.tsx    User settings UI
  |     |
  |     +-- terminal/
  |     |     +-- TerminalPanel.tsx  xterm.js with WebSocket + local fallback
  |     |
  |     +-- common/
  |           +-- CommandPalette.tsx  File search + command mode
  |           +-- ErrorBoundary.tsx   React error boundary with fallback
  |
  +-- services/
  |     +-- api.ts              REST client (auth tokens from localStorage)
  |     +-- git.ts              Git API client
  |     +-- websocket.ts        TerminalWebSocket class with reconnect
  |
  +-- hooks/
  |     +-- useKeyboardShortcuts.ts  Global Ctrl+P/B/`/W/Shift+E/F/G, Ctrl+1-9
  |     +-- useFileLanguage.ts       Extension-to-Monaco language mapping
  |
  +-- types/
        +-- index.ts            Shared type definitions
```

---

## 4. Data Flow Diagrams `[ARCH-004]`

### 4.1 File Open Flow

```
User clicks file in FileExplorer
  |
  v
dispatch(OPEN_FILE, { tab with empty content })
  |
  v
readFile(sessionId, path) --HTTP GET--> /api/v1/workspaces/{sid}/files/content
  |                                              |
  v                                              v
dispatch(SET_TAB_CONTENT, { content })    FileTreeHandler.HandleReadFile()
  |                                              |
  v                                              v
Monaco Editor renders content              os.ReadFile(path) + size check
```

### 4.2 File Save Flow

```
User presses Ctrl+S in CodeEditor
  |
  v
writeFile(sessionId, path, content) --HTTP PUT--> /api/v1/workspaces/{sid}/files/content
  |                                                        |
  v                                                        v
dispatch(MARK_TAB_SAVED)                           FileTreeHandler.HandleWriteFile()
  |                                                        |
  v                                                        v
Tab dot indicator clears                           os.WriteFile(path, content)
```

### 4.3 Terminal Session Flow

```
User opens terminal
  |
  v
TerminalWebSocket.connect() --WS UPGRADE--> /ws/terminal?session_id=X
  |                                                  |
  v                                                  v
xterm.js initialized                         Handler.ServeHTTP()
  |                                                  |
  |                                          Upgrader.Upgrade(w, r, nil)
  |                                                  |
  |                                          container.AttachToContainer(sessionID)
  |                                                  |
  |                                          Docker exec create + attach
  |                                                  |
  v                                                  v
User types keystroke ----binary----> ReadPump -> containerWriter -> bash stdin
  |                                                  |
  v                                                  v
xterm.js renders    <----binary---- WritePump <- send channel <- streamContainerOutput <- bash stdout
```

### 4.4 Workspace Creation Flow

```
POST /api/v1/workspaces (with JWT)
  |
  v
Auth middleware extracts userID from JWT
  |
  v
ContainerManager.CreateWorkspace(ctx, userID)
  |
  +-- Generate UUID session ID
  +-- Build container config (image, memory, CPU, PID, security)
  +-- docker.ContainerCreate(config)
  +-- docker.ContainerStart(containerID)
  +-- Store session in sessions map (with CreatedAt)
  |
  v
Return { session_id, container_id, status: "running" }
```

### 4.5 Graceful Shutdown Flow

```
SIGINT or SIGTERM received
  |
  v
server.Shutdown(ctx)          -- Drain HTTP connections (30s timeout)
  |
  v
hub.Stop()                    -- Close done channel, disconnect all WS clients
  |
  v
containerMgr.Shutdown(ctx)    -- Stop all running containers (force remove)
  |
  v
rateLimiter.Stop()            -- Stop cleanup goroutine
  |
  v
Process exits
```

---

## 5. Technology Stack `[ARCH-005]`

### 5.1 Backend

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Language | Go | 1.22 | Backend server |
| Container SDK | docker/docker | v25.0.3 | Container lifecycle management |
| WebSocket | gorilla/websocket | v1.5.1 | Terminal bi-directional I/O |
| UUID | google/uuid | v1.6.0 | Session ID and request ID generation |
| Auth | HMAC-SHA256 | — | JWT token validation |

### 5.2 Frontend

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Framework | React | 19.2.4 | UI components |
| Language | TypeScript | 5.9.3 | Type-safe JavaScript |
| Editor | Monaco Editor | 4.7.0 via @monaco-editor/react | Code editing |
| Terminal | xterm.js | 6.0.0 via @xterm/xterm | Terminal emulation |
| Styling | Tailwind CSS | 4.2.2 | Utility-first CSS |
| Icons | lucide-react | 1.0.1 | UI icons |
| Routing | react-router-dom | 7.13.2 | Client-side routing |
| Build | Vite | 8.0.1 | Development server and bundler |

### 5.3 Testing

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Go tests | testing + race detector | built-in | Backend unit tests |
| Frontend tests | Vitest | 4.1.2 | Frontend unit/component tests |
| DOM testing | @testing-library/react | 16.3.2 | React component testing |
| DOM environment | jsdom | 29.0.1 | Browser-like test environment |

### 5.4 Infrastructure

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Container runtime | Docker | Workspace isolation |
| Orchestration | docker-compose | Local development stack |
| CI/CD | GitHub Actions | Automated build/test pipeline |
| API image | Alpine 3.19 (multi-stage) | Minimal production image |
| Workspace image | Ubuntu 22.04 | Development environment |

---

## 6. Security Architecture `[ARCH-006]`

### 6.1 Authentication Layer

```
Client --Bearer JWT--> Auth Middleware --user_id/email in context--> Handlers
```

- HMAC-SHA256 signature validation
- Expiration enforcement
- User identity injected into request context
- `/health` exempt from authentication

### 6.2 Container Isolation

| Control | Implementation |
|---------|---------------|
| Memory limit | Configurable via `CONTAINER_MEMORY_MB` |
| CPU limit | Configurable via `CONTAINER_CPU_SHARES` |
| PID limit | Hard-coded 512 |
| Privilege escalation | `no-new-privileges` security option |
| Network | Isolated bridge network (`cloudide-net`) |
| Filesystem | Writable workspace volume at `/workspace` |
| Lifetime | Enforced via cleanup loop (`CONTAINER_TIMEOUT_MIN`) |

### 6.3 Input Validation

| Protection | Implementation |
|------------|---------------|
| Path traversal | `filepath.Clean()` + prefix check against workspace base path |
| File size | 10MB read limit, 10MB request body limit |
| Rate limiting | Token-bucket per-IP (configurable RPS + burst) |
| CORS | Origin whitelist from `ALLOWED_ORIGINS` |
| WebSocket origin | Same origin check against `ALLOWED_ORIGINS` |

### 6.4 Secret Management

| Secret | Storage | Notes |
|--------|---------|-------|
| `JWT_SECRET` | Environment variable | Required, no default. Docker Compose reads from `.env` file |
| Auth tokens (frontend) | `localStorage` under `cloudcode_token` | Cleared on logout |

---

## 7. Scalability Considerations `[ARCH-007]`

### Current Architecture (Single Node)

The current architecture runs all components on a single host:
- One Go process handles all HTTP and WebSocket connections
- All workspace containers run on the same Docker engine
- State is in-memory (no database)

### Scaling Path

| Constraint | Current Limit | Scaling Strategy |
|------------|--------------|------------------|
| Concurrent workspaces | ~50 per host (4GB each) | Multi-node with container orchestration (K8s) |
| WebSocket connections | ~10,000 per Go process | Horizontal scaling with sticky sessions |
| File operations | Local filesystem | Network-attached storage or object storage |
| Session state | In-memory map | Redis or database-backed sessions |
| Auth | Single JWT secret | JWK key rotation, external IdP (OAuth2) |

---

## 8. Network Architecture `[ARCH-008]`

```
Internet
  |
  | HTTPS (TLS terminated at load balancer or directly)
  v
+------------------+
|  Go HTTP Server  |  Port 8080 (configurable)
|                  |
|  /health         |  Unauthenticated health probe
|  /api/v1/*       |  Authenticated REST API
|  /ws/terminal    |  Authenticated WebSocket
+--------+---------+
         |
         | Docker API (unix socket)
         v
+------------------+
|  Docker Engine   |
|                  |
|  Bridge Network: cloudide-net
|  Containers communicate via network name
+------------------+
```

**Frontend development:** Vite dev server on port 3000 proxies `/api/*` and `/ws/*` to backend on port 8080.

**Production:** Frontend is a static build served by a CDN or nginx, with API calls routed to the Go backend.

---

## Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-03-31 | CloudCode Team | Initial specification |
