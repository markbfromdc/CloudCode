# API Specification

> **Version:** 1.0.0
> **Last Updated:** 2026-03-31
> **Status:** Approved
> **Base URL:** `http://localhost:8080` (development), configurable via `HTTP_PORT` and `HOST_ADDR`

---

## 1. Authentication

### 1.1 Method — JWT Bearer Token `[API-AUTH-001]`

All endpoints except `/health` require a valid JWT in the `Authorization` header.

```
Authorization: Bearer <token>
```

**Token format:** HMAC-SHA256 signed JWT (HS256)

**Claims payload:**

| Claim | Type | Required | Description |
|-------|------|----------|-------------|
| `sub` | string | Yes | User ID |
| `email` | string | Yes | User email address |
| `exp` | int64 | Yes | Expiration (Unix timestamp) |
| `iat` | int64 | Yes | Issued at (Unix timestamp) |

**Validation rules:**
- Signature verified against `JWT_SECRET` using HMAC-SHA256
- `exp` must be in the future
- Token format must be `header.payload.signature` (3 dot-separated parts)

**Context injection:** On successful validation, `sub` is injected as `user_id` and `email` as `email` into the request context for downstream handlers.

**Error responses:**

| Condition | Status | Body |
|-----------|--------|------|
| Missing `Authorization` header | 401 | `{"error":"missing authorization header"}` |
| Invalid format (not `Bearer <token>`) | 401 | `{"error":"invalid authorization format"}` |
| Invalid signature | 401 | `{"error":"invalid token"}` |
| Expired token | 401 | `{"error":"invalid token"}` |
| Malformed payload | 401 | `{"error":"invalid token"}` |

### 1.2 Frontend Token Storage `[API-AUTH-002]`

The frontend reads the JWT from `localStorage` under key `cloudcode_token` and includes it in all API and WebSocket requests.

---

## 2. Rate Limiting `[API-RATE-001]`

All endpoints (including `/health`) are subject to per-IP token-bucket rate limiting.

| Parameter | Default | Env Var |
|-----------|---------|---------|
| Requests per second | 100 | `RATE_LIMIT_RPS` |
| Burst size | 200 | `RATE_LIMIT_BURST` |

**Response headers:**

| Header | Description |
|--------|-------------|
| `X-RateLimit-Remaining` | Tokens remaining in the bucket |

**When exceeded:**

```
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{"error":"rate limit exceeded"}
```

**IP extraction priority:** `X-Forwarded-For` header, then `X-Real-IP`, then `RemoteAddr`.

---

## 3. Request Correlation `[API-REQ-001]`

Every response includes an `X-Request-ID` header containing a UUID v4. If the incoming request already carries `X-Request-ID`, the server uses that value instead of generating a new one.

The request ID is available in the request context for logging and error responses.

---

## 4. Middleware Chain `[API-MW-001]`

Requests are processed through this middleware stack in order:

```
Client → RequestLogger → RateLimiter → RequestID → CORS → [Auth] → Handler
```

`/health` bypasses the `Auth` middleware. All other routes pass through `Auth`.

---

## 5. Endpoints

### 5.1 Health Check

#### `GET /health` `[API-001]`

Unauthenticated. Returns server health status.

**Response 200:**
```json
{
  "status": "healthy",
  "active_sessions": 2,
  "active_workspaces": 2,
  "timestamp": "2026-03-31T12:00:00Z"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Always `"healthy"` if server is responding |
| `active_sessions` | int | Number of WebSocket sessions in the hub |
| `active_workspaces` | int | Number of running workspace containers |
| `timestamp` | string | ISO 8601 UTC timestamp |

---

### 5.2 Workspace Management

#### `POST /api/v1/workspaces` `[API-002]`

Create a new isolated workspace container for the authenticated user.

**Request:** Empty body. User identity extracted from JWT claims.

**Response 201:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "container_id": "a1b2c3d4e5f6...",
  "status": "running"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | UUID identifying this workspace session |
| `container_id` | string | Docker container ID (full SHA) |
| `status` | string | Container status: `"running"` |

**Error responses:**

| Condition | Status | Body |
|-----------|--------|------|
| Not POST | 405 | `{"error":"method not allowed"}` |
| Missing user context | 401 | `{"error":"unauthorized"}` |
| Docker create fails | 500 | `{"error":"failed to create workspace"}` |

**Container configuration applied:**
- Image: `WORKSPACE_IMAGE` (default: `cloudide-workspace:latest`)
- Memory limit: `CONTAINER_MEMORY_MB` MB
- CPU shares: `CONTAINER_CPU_SHARES`
- PID limit: 512
- Security: `no-new-privileges`
- Network: `DOCKER_NETWORK`
- Working directory: `/workspace`
- Environment: `SESSION_ID`, `USER_ID` injected

---

#### `POST /api/v1/workspaces/stop` `[API-003]`

Stop and remove a workspace container.

**Query parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `session_id` | string | Yes | Session ID to stop |

**Response 200:**
```json
{
  "status": "stopped"
}
```

**Error responses:**

| Condition | Status | Body |
|-----------|--------|------|
| Missing `session_id` | 400 | `{"error":"session_id required"}` |
| Session not found | 500 | `{"error":"failed to stop workspace"}` |

---

### 5.3 File Tree API

#### `GET /api/v1/workspaces/{sessionId}/files` `[API-004]`

List files and directories in the workspace.

**Query parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `path` | string | `/` | Directory path to list |

**Response 200:**
```json
[
  {
    "name": "src",
    "path": "/workspace/src",
    "type": "directory",
    "children": [
      {
        "name": "main.ts",
        "path": "/workspace/src/main.ts",
        "type": "file"
      }
    ]
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | File or directory name |
| `path` | string | Absolute path |
| `type` | string | `"file"` or `"directory"` |
| `children` | FileNode[] | Nested entries (directories only) |

**Constraints:**
- Maximum tree depth: 3 levels
- Hidden files (starting with `.`) are excluded
- Directories `node_modules`, `__pycache__`, `vendor` are excluded

---

#### `GET /api/v1/workspaces/{sessionId}/files/content` `[API-005]`

Read file content.

**Query parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | Yes | Absolute file path |

**Response 200:** Raw file content as `text/plain`.

**Error responses:**

| Condition | Status | Body |
|-----------|--------|------|
| File not found | 404 | `{"error":"file not found"}` |
| File > 10MB | 413 | `{"error":"file too large"}` |
| Path outside workspace | 403 | `{"error":"access denied"}` |

---

#### `PUT /api/v1/workspaces/{sessionId}/files/content` `[API-006]`

Write file content.

**Query parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | Yes | Absolute file path |

**Request body:**
```json
{
  "content": "file content here"
}
```

**Response 200:**
```json
{
  "status": "ok"
}
```

**Constraints:**
- Request body limit: 10MB
- Parent directories created automatically
- Path traversal blocked (must be under workspace base path)

---

### 5.4 File Operations API

#### `POST /api/v1/files/create` `[API-007]`

Create a file or directory.

**Request body:**
```json
{
  "path": "/workspace/src/newfile.ts",
  "type": "file"
}
```

| Field | Type | Required | Values | Description |
|-------|------|----------|--------|-------------|
| `path` | string | Yes | — | Absolute path |
| `type` | string | Yes | `"file"`, `"directory"` | What to create |

**Response 201:**
```json
{
  "status": "created",
  "path": "/workspace/src/newfile.ts",
  "type": "file"
}
```

**Error responses:**

| Condition | Status | Body |
|-----------|--------|------|
| Empty path | 400 | `{"error":"path required"}` |
| Path outside workspace | 403 | `{"error":"access denied"}` |

---

#### `DELETE /api/v1/files/delete` `[API-008]`

Delete a file or directory.

**Query parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | Yes | Absolute path to delete |

**Response 200:**
```json
{
  "status": "deleted",
  "path": "/workspace/src/oldfile.ts"
}
```

**Error responses:**

| Condition | Status | Body |
|-----------|--------|------|
| Missing path | 400 | `{"error":"path required"}` |
| Path outside workspace | 403 | `{"error":"access denied"}` |
| Deleting workspace root | 403 | `{"error":"cannot delete workspace root"}` |

---

#### `POST /api/v1/files/rename` `[API-009]`

Rename or move a file/directory.

**Request body:**
```json
{
  "old_path": "/workspace/src/old.ts",
  "new_path": "/workspace/src/new.ts"
}
```

**Response 200:**
```json
{
  "status": "renamed",
  "old_path": "/workspace/src/old.ts",
  "new_path": "/workspace/src/new.ts"
}
```

---

### 5.5 Git Operations API

All git endpoints accept an optional `workspace` query parameter (defaults to `/workspace`).

#### `GET /api/v1/git/status` `[API-010]`

**Response 200:**
```json
[
  {
    "path": "main.ts",
    "status": "modified",
    "status_code": "M"
  }
]
```

| `status` values | `status_code` |
|-----------------|---------------|
| `added` | `A` |
| `modified` | `M` |
| `deleted` | `D` |
| `renamed` | `R` |
| `untracked` | `??` |

---

#### `GET /api/v1/git/log` `[API-011]`

Returns last 20 commits.

**Response 200:**
```json
[
  {
    "hash": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
    "author": "Developer Name",
    "date": "2026-03-31",
    "message": "Fix authentication bug"
  }
]
```

---

#### `GET /api/v1/git/branches` `[API-012]`

**Response 200:**
```json
[
  { "name": "main", "current": true },
  { "name": "feature/auth", "current": false }
]
```

---

#### `POST /api/v1/git/commit` `[API-013]`

**Request body:**
```json
{
  "message": "Fix bug in auth module",
  "files": ["src/auth.ts"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message` | string | Yes | Commit message |
| `files` | string[] | No | Files to stage (omit for `git add -A`) |

**Response 200:**
```json
{
  "status": "committed",
  "output": "[main abc1234] Fix bug in auth module\n 1 file changed"
}
```

**Error responses:**

| Condition | Status | Body |
|-----------|--------|------|
| Empty message | 400 | `{"error":"commit message required"}` |
| Commit fails | 500 | `{"error":"commit failed: <git output>"}` |

Note: Commits use `--no-gpg-sign` as workspace containers lack GPG keys.

---

#### `POST /api/v1/git/stage` `[API-014]`

**Request body:**
```json
{
  "files": ["src/main.ts", "src/app.ts"]
}
```

Empty `files` array stages all changes (`git add -A`).

**Response 200:**
```json
{
  "status": "staged"
}
```

---

#### `POST /api/v1/git/init` `[API-015]`

Initialize a new git repository in the workspace.

**Response 200:**
```json
{
  "status": "initialized"
}
```

Automatically configures `user.email` as `developer@cloudcode.dev` and `user.name` as `CloudCode Developer`.

---

### 5.6 WebSocket Terminal

#### `WS /ws/terminal` `[API-016]`

Upgrades to WebSocket for bidirectional terminal I/O.

**Query parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `session_id` | string | Yes | Workspace session ID |

**Authentication:** JWT validated from request context (set by auth middleware).

**WebSocket configuration:**
- Read buffer: `WS_READ_BUFFER_SIZE` bytes (default 8192)
- Write buffer: `WS_WRITE_BUFFER_SIZE` bytes (default 8192)
- Origin check: Must match `ALLOWED_ORIGINS`
- Binary type: `arraybuffer`

**Message flow:**
- **Client → Server:** Raw keystrokes (binary) forwarded to container stdin
- **Server → Client:** Container stdout output (binary) streamed via send buffer

**Heartbeat:**
- Ping interval: `WS_PING_INTERVAL_SEC` seconds (default 30)
- Pong timeout: `WS_PONG_TIMEOUT_SEC` seconds (default 40)
- Max message size: `WS_MAX_MESSAGE_SIZE` bytes (default 65536)
- Write timeout: `WS_WRITE_TIMEOUT_SEC` seconds (default 10)
- Send buffer: 256 messages (drops with warning when full)

**Resize message (client → server):**
```json
{
  "type": "resize",
  "cols": 120,
  "rows": 40
}
```

**Connection lifecycle:**
1. Client sends HTTP upgrade with `session_id` and JWT
2. Server validates auth and creates Docker exec session (bash shell)
3. ReadPump goroutine forwards keystrokes to container stdin
4. WritePump goroutine sends container stdout to client
5. StreamContainerOutput goroutine reads container stdout into send buffer
6. On disconnect: cancel context, close container writer, unregister from hub

**Reconnection (client-side):**
- Exponential backoff: `1000ms * 2^attempt`
- Maximum attempts: 5
- Resets attempt counter on successful connection

---

## 6. Error Response Format `[API-ERR-001]`

All error responses use this JSON structure:

```json
{
  "code": 404,
  "message": "file not found",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `code` | int | HTTP status code |
| `message` | string | Human-readable error description |
| `request_id` | string | Correlation ID (omitted if unavailable) |

### Standard Error Codes

| Status | Meaning | When |
|--------|---------|------|
| 400 | Bad Request | Missing required parameters, invalid JSON |
| 401 | Unauthorized | Missing or invalid JWT |
| 403 | Forbidden | Path traversal attempt, workspace root deletion |
| 404 | Not Found | File or session not found |
| 405 | Method Not Allowed | Wrong HTTP method for endpoint |
| 413 | Request Entity Too Large | File exceeds 10MB limit |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Docker failures, git errors, server bugs |

---

## 7. CORS Configuration `[API-CORS-001]`

| Header | Value |
|--------|-------|
| `Access-Control-Allow-Origin` | Matched from `ALLOWED_ORIGINS` |
| `Access-Control-Allow-Methods` | `GET, POST, PUT, DELETE, OPTIONS` |
| `Access-Control-Allow-Headers` | `Content-Type, Authorization` |
| `Access-Control-Max-Age` | `86400` (24 hours) |

Preflight `OPTIONS` requests return 204 with the above headers.
Origins not in the allowed list receive no CORS headers (browser blocks the request).

---

## Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-03-31 | CloudCode Team | Initial specification |
