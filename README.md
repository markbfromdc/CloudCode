# CloudCode

A cloud-based integrated development environment that provisions isolated Docker containers per workspace, giving each user a full Linux environment with editor, terminal, git, and file management — accessible from the browser.

## Architecture

```
Browser (React SPA)
    |
    +-- REST API --> Go HTTP Server (:8080)
    |                   +-- JWT Auth + Rate Limiting
    |                   +-- File Tree / File Ops / Git API
    |                   +-- Container Manager (Docker SDK)
    |                          +-- Isolated workspace containers
    |
    +-- WebSocket --> /ws/terminal
                        +-- Docker exec attach (bidirectional I/O)
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22, Docker SDK v25, gorilla/websocket |
| Frontend | React 19, TypeScript 5.9, Monaco Editor, xterm.js |
| Styling | Tailwind CSS 4.2 |
| Build | Vite 8, Docker multi-stage |
| Testing | Go test (race), Vitest, @testing-library/react |

## Quick Start

### Prerequisites
- Go 1.22+
- Node.js 20+
- Docker

### Development

```bash
# Backend
cp .env.example .env
# Edit .env and set JWT_SECRET
make run

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

### Docker

```bash
make docker-up    # Build and start all services
make docker-down  # Stop and clean up
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check (unauthenticated) |
| POST | /api/v1/workspaces | Create workspace |
| POST | /api/v1/workspaces/stop | Stop workspace |
| GET | /api/v1/workspaces/{sid}/files | List file tree |
| GET | /api/v1/workspaces/{sid}/files/content | Read file |
| PUT | /api/v1/workspaces/{sid}/files/content | Write file |
| POST | /api/v1/files/create | Create file/directory |
| DELETE | /api/v1/files/delete | Delete file/directory |
| POST | /api/v1/files/rename | Rename/move file |
| GET | /api/v1/git/status | Git status |
| GET | /api/v1/git/log | Commit history |
| GET | /api/v1/git/branches | Branch list |
| POST | /api/v1/git/commit | Create commit |
| POST | /api/v1/git/stage | Stage files |
| POST | /api/v1/git/init | Initialize repo |
| WS | /ws/terminal | Terminal WebSocket |

## Testing

```bash
make test           # Go tests with race detector
make test-frontend  # Vitest frontend tests
make test-all       # Both
make coverage       # Generate coverage report
```

## Configuration

See [.env.example](.env.example) for all environment variables.

## License

MIT
