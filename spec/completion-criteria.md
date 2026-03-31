# Completion Criteria Specification

> **Version:** 1.0.0
> **Last Updated:** 2026-03-31
> **Status:** Approved

---

## 1. Overview `[EVAL-001]`

This document defines the criteria, weights, and methodology used by the project completion evaluation system (`cmd/evaluate`) to compute an overall completion percentage for the CloudCode project.

The system produces a weighted score from 0-100% across 7 categories, each containing individual pass/fail/partial checks.

---

## 2. Scoring Methodology `[EVAL-002]`

### 2.1 Formula

```
OverallScore = Σ (CategoryPercentage × CategoryWeight)
```

Where `CategoryPercentage` is computed as:
```
CategoryPercentage = (Σ ItemScore / Σ ItemMaxScore) × 100
```

Each item has a normalized score (0-1) and max score (typically 1). Partial scores are supported (e.g., 12/16 endpoints = 0.75).

### 2.2 Thresholds

| Score Range | Rating | Description |
|-------------|--------|-------------|
| 95-100% | Production Ready | All criteria met, ready for deployment |
| 85-94% | Near Complete | Minor gaps, safe for staging |
| 70-84% | In Progress | Core functionality works, gaps remain |
| 50-69% | Early Development | Foundational work done, major gaps |
| 0-49% | Initial | Scaffolding only |

---

## 3. Categories and Weights `[EVAL-003]`

| Category | Weight | Description |
|----------|--------|-------------|
| Code Implementation | 25% | All planned packages, endpoints, and components exist |
| Test Coverage | 25% | Test files exist, meet minimum counts, suites pass |
| Spec Compliance | 15% | Requirement IDs covered, acceptance criteria met |
| Build & CI | 10% | Go/TS/Vite builds succeed, CI pipeline exists |
| Documentation | 10% | README, .env.example, specs, doc comments |
| Security | 10% | Auth, rate limiting, CORS, path traversal, container security |
| Infrastructure | 5% | Dockerfiles, Compose, Makefile targets |

Weights sum to 100%.

---

## 4. Category Details

### 4.1 Code Implementation (25%) `[EVAL-010]`

| Check | Max Score | Pass Condition |
|-------|-----------|----------------|
| Backend package: internal/api | 1 | Directory exists |
| Backend package: internal/config | 1 | Directory exists |
| Backend package: internal/container | 1 | Directory exists |
| Backend package: internal/middleware | 1 | Directory exists |
| Backend package: internal/websocket | 1 | Directory exists |
| Backend package: internal/logging | 1 | Directory exists |
| Server entry: cmd/server/main.go | 1 | File exists |
| Frontend dir: src/components | 1 | Directory exists |
| Frontend dir: src/services | 1 | Directory exists |
| Frontend dir: src/hooks | 1 | Directory exists |
| Frontend dir: src/context | 1 | Directory exists |
| Frontend dir: src/types | 1 | Directory exists |
| API endpoint registrations | 1 | >= 16 HandleFunc/Handle calls in main.go |
| Frontend components | 1 | >= 18 .tsx files in components/ |

### 4.2 Test Coverage (25%) `[EVAL-020]`

| Check | Max Score | Pass Condition |
|-------|-----------|----------------|
| Tests: internal/api | 1 | >= 40 test functions |
| Tests: internal/container | 1 | >= 15 test functions |
| Tests: internal/middleware | 1 | >= 10 test functions |
| Tests: internal/websocket | 1 | >= 10 test functions |
| Tests: internal/config | 1 | >= 5 test functions |
| Tests: internal/logging | 1 | >= 5 test functions |
| Frontend test: WorkspaceContext | 1 | File exists |
| Frontend test: api.test.ts | 1 | File exists |
| Frontend test: git.test.ts | 1 | File exists |
| Frontend test: websocket.test.ts | 1 | File exists |
| Frontend test: useFileLanguage.test.ts | 1 | File exists |
| Go test suite | 1 | All tests pass with -race |
| Frontend test suite | 1 | All tests pass |

### 4.3 Spec Compliance (15%) `[EVAL-030]`

| Check | Max Score | Pass Condition |
|-------|-----------|----------------|
| Traceability matrix entries | 1 | >= 5 requirement IDs in spec/README.md |
| Spec IDs with test coverage | 1 | >= 80% of [AREA-NNN] IDs appear in test docs |
| User story acceptance criteria | 1 | >= 90% of criteria met across all stories |

### 4.4 Build & CI (10%) `[EVAL-040]`

| Check | Max Score | Pass Condition |
|-------|-----------|----------------|
| Go build | 1 | `go build ./...` exits 0 |
| TypeScript check | 1 | `npx tsc --noEmit` exits 0 |
| Frontend production build | 1 | `npx vite build` exits 0 |
| CI pipeline exists | 1 | .github/workflows/ci.yml exists |
| CI covers backend + frontend | 1 | CI file contains Go and Node.js steps |

### 4.5 Documentation (10%) `[EVAL-050]`

| Check | Max Score | Pass Condition |
|-------|-----------|----------------|
| Root README.md | 1 | File exists |
| .env.example | 1 | File exists |
| Spec: README.md | 1 | File exists |
| Spec: api-specification.md | 1 | File exists |
| Spec: system-architecture.md | 1 | File exists |
| Spec: database-schema.md | 1 | File exists |
| Spec: user-stories.md | 1 | File exists |
| Spec: testing-specification.md | 1 | File exists |
| Spec: deployment-specification.md | 1 | File exists |
| Go package doc comments | 1 | All 6 packages have `// Package` comments |

### 4.6 Security (10%) `[EVAL-060]`

| Check | Max Score | Pass Condition |
|-------|-----------|----------------|
| JWT authentication middleware | 1 | internal/middleware/auth.go exists |
| Rate limiting middleware | 1 | internal/middleware/ratelimit.go exists |
| Request correlation middleware | 1 | internal/middleware/requestid.go exists |
| CORS middleware | 1 | internal/middleware/cors.go exists |
| Path traversal protection | 1 | >= 2 files in api/ use filepath.Clean |
| Container security options | 1 | no-new-privileges in container/manager.go |

### 4.7 Infrastructure (5%) `[EVAL-070]`

| Check | Max Score | Pass Condition |
|-------|-----------|----------------|
| API Dockerfile | 1 | Dockerfile.api exists |
| Workspace Dockerfile | 1 | workspace/Dockerfile exists |
| Docker Compose config | 1 | docker-compose.yml exists |
| Makefile | 1 | Makefile exists |
| Makefile targets | 1 | build, test, run, docker-build, docker-up, clean targets present |

---

## 5. Milestone Tracking `[EVAL-080]`

Milestones are extracted from `spec/user-stories.md`. Each user story (US-001 through US-016) is a milestone with:

- **ID:** From the story header (e.g., `US-001`)
- **Priority:** P0 (critical), P1 (high), P2 (medium)
- **Criteria count:** Number of `- [ ]` and `- [x]` lines
- **Completed count:** Number of `- [x]` lines
- **Status:** `complete` (all criteria checked), `partial` (some checked), `incomplete` (none checked)
- **Dependencies:** Extracted from "Traceability" lines

---

## 6. CLI Usage `[EVAL-090]`

```bash
# Terminal report
go run ./cmd/evaluate -dir . -format terminal

# JSON report
go run ./cmd/evaluate -dir . -format json -output report.json

# Both formats
go run ./cmd/evaluate -dir . -format both -output report.json

# Skip slow checks (builds and test execution)
go run ./cmd/evaluate -dir . -skip-build -skip-tests

# Via Makefile
make evaluate
```

---

## 7. Report Format `[EVAL-100]`

### 7.1 JSON Schema

```json
{
  "timestamp": "2026-03-31T12:00:00Z",
  "overall_score": 94.2,
  "categories": [
    {
      "name": "Code Implementation",
      "weight": 0.25,
      "score": 13.0,
      "percentage": 92.9,
      "items": [
        {
          "name": "Backend package: internal/api",
          "status": "pass",
          "score": 1.0,
          "max_score": 1.0,
          "detail": "Directory exists"
        }
      ]
    }
  ],
  "milestones": [
    {
      "id": "US-001",
      "name": "Create Workspace",
      "priority": "P0",
      "status": "complete",
      "criteria": 7,
      "completed": 7,
      "depends_on": ["API-002", "ARCH-001"]
    }
  ],
  "summary": {
    "total_files": 63,
    "total_loc": 7900,
    "go_tests": 109,
    "frontend_tests": 86,
    "passing_tests": 195,
    "failing_tests": 0,
    "build_status": "pass",
    "typecheck_status": "pass"
  }
}
```

---

## Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-03-31 | CloudCode Team | Initial specification |
