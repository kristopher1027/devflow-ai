# DevFlow AI

**AI-powered engineering workspace for modern software development.**

DevFlow AI is a full-stack developer productivity platform designed to bring projects, tasks, GitHub activity, code intelligence, documentation, CI/CD insights, and AI-assisted engineering into one workspace.

The goal is to help developers understand and manage the complete software development lifecycle from a single platform.

---

## 🚀 Vision

Modern development workflows are spread across many tools:

- GitHub for repositories, issues, and pull requests
- Project-management tools for tasks
- CI/CD platforms for build and deployment information
- Documentation platforms for technical knowledge
- AI tools for coding assistance
- Chat applications for team communication

**DevFlow AI aims to connect these pieces into one engineering workspace.**

The long-term vision is for DevFlow to understand a project's:

> **Code + Repository + Tasks + Issues + Pull Requests + CI/CD + Documentation + Development History**

and provide useful context-aware engineering assistance.

---

## ✨ Planned Features

### Workspace Management

- Create and manage engineering workspaces
- Workspace ownership and membership
- Member roles and permissions
- Secure workspace-level access control

### Project & Task Management

- Create and organize projects
- Track engineering tasks
- Task status and priorities
- Project activity and history

### GitHub Integration

- Connect GitHub repositories
- Import repositories and project information
- View issues and pull requests
- Track repository activity
- Connect development activity with project tasks

### Code Intelligence

- Understand repository structure
- Index source code and documentation
- Search code using natural language
- Retrieve relevant code context
- Build project-aware AI context using RAG

### AI Engineering Assistant

DevFlow's AI assistant is intended to help developers:

- Understand unfamiliar code
- Explain errors
- Investigate bugs
- Suggest implementation approaches
- Analyze technical documentation
- Understand pull requests
- Generate development insights
- Answer questions about a connected codebase

### CI/CD Intelligence

- Monitor build failures
- Analyze CI/CD errors
- Connect failures to relevant code
- Provide debugging context
- Track deployment-related information

### Observability

Planned production capabilities include:

- Structured logging
- Metrics
- Distributed tracing
- Error tracking
- Health checks
- Service monitoring

---

## 🏗️ Architecture

DevFlow is being developed using a modular full-stack architecture.

```text
┌─────────────────────────────────────────────┐
│                 Frontend                    │
│          Modern Web Application             │
└──────────────────────┬──────────────────────┘
                       │
                       │ HTTP / REST
                       ▼
┌─────────────────────────────────────────────┐
│                 Go API                      │
│                                             │
│  HTTP → Service → Repository → Database     │
└──────────────────────┬──────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────┐
│               PostgreSQL                    │
│                                             │
│ Users • Workspaces • Projects • Tasks       │
└─────────────────────────────────────────────┘

              External Integrations
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
       GitHub         AI          CI/CD
```

The backend follows a layered architecture:

```text
HTTP Handlers
      ↓
Services
      ↓
Repositories
      ↓
PostgreSQL
```

This separation keeps business logic independent from HTTP and database implementations, making the system easier to test, maintain, and scale.

---

## 🛠️ Technology Stack

### Backend

- **Go**
- REST API
- PostgreSQL
- SQL migrations
- Layered architecture
- Unit and integration testing

### Database

- **PostgreSQL 17**
- UUID-based identifiers
- Database migrations
- Foreign-key relationships
- Transaction-safe operations

### Frontend

The frontend will use a modern TypeScript-based web stack and communicate with the Go API through REST APIs.

### AI

The AI layer is planned to incorporate:

- LLM-based assistance
- Retrieval-Augmented Generation (RAG)
- Code and documentation indexing
- Repository-aware context
- Embeddings/vector search

### Infrastructure

Planned infrastructure includes:

- Docker
- Docker Compose
- CI/CD
- Structured logging
- Metrics
- Distributed tracing
- Production deployment

---

## 📁 Project Structure

The project is organized as a monorepo:

```text
devflow-ai/
├── apps/
│   └── api/
│       ├── cmd/
│       │   └── server/
│       ├── internal/
│       │   ├── config/
│       │   ├── database/
│       │   ├── domain/
│       │   ├── http/
│       │   ├── repository/
│       │   ├── server/
│       │   └── service/
│       ├── migrations/
│       ├── go.mod
│       └── go.sum
│
├── docker-compose.yml
├── README.md
└── ...
```

The exact structure will evolve as additional services and frontend applications are introduced.

---

## 🔐 Security

Security is a first-class requirement of DevFlow.

The application is being designed around principles including:

- Authentication and authorization
- Workspace-level access control
- Role-based permissions
- Input validation
- Secure password handling
- SQL injection prevention
- Secure session/token management
- Least-privilege access
- Secret management
- Secure GitHub integration
- Protection of sensitive repository data

Security considerations will be incorporated throughout development rather than added after the application is completed.

---

## 🧪 Testing

DevFlow follows a test-first approach where practical.

Current backend testing includes:

- Unit tests
- Service-layer tests
- Repository tests
- HTTP handler tests
- Database-backed integration tests

Run the Go test suite with:

```bash
cd apps/api
DATABASE_URL=postgres://devflow:devflowpassword@localhost:5432/devflow?sslmode=disable go test ./...
```

---

## 🐳 Running Locally

### Prerequisites

Make sure you have:

- Go
- Docker
- Docker Compose
- PostgreSQL-compatible environment
- Git

### Clone the repository

```bash
git clone https://github.com/kristopher1027/devflow-ai.git
cd devflow-ai
```

### Start PostgreSQL

```bash
docker compose up -d
```

### Configure environment variables

Create the appropriate environment configuration for the API.

Example:

```env
DATABASE_URL=postgres://devflow:devflowpassword@localhost:5432/devflow?sslmode=disable
PORT=8080
```

> Do not commit real credentials, API keys, tokens, or secrets to Git.

### Run migrations

Using the project's migration tooling:

```bash
migrate -path migrations \
  -database "$DATABASE_URL" \
  up
```

### Start the API

```bash
cd apps/api
go run ./cmd/server
```

The API should be available at:

```text
http://localhost:8080
```

---

## 🩺 Health Check

The API provides a health endpoint for verifying that the service is running.

Example response:

```json
{
  "service": "devflow-api",
  "status": "ok"
}
```

---

## 🗄️ Current Backend Progress

DevFlow is being developed incrementally rather than generating the entire application at once.

Current foundation includes:

- Go API
- PostgreSQL integration
- Database configuration
- Database migrations
- User domain model
- User repository
- Workspace domain model
- Workspace persistence
- Workspace ownership
- Workspace membership/roles
- Service layer
- HTTP handlers
- Automated tests
- Docker-based PostgreSQL development environment

The next capabilities will build on this foundation.

### Working Change Log

This section is the handoff record for ongoing development. Each future change should update this README with:

- What was added or changed
- The relevant API, domain, database, or infrastructure files
- Tests added or updated
- Validation performed
- Remaining work or the next recommended step

#### Baseline: 2026-09-17

Already present in the repository:

- Authentication helpers for passwords, sessions, tokens, and token hashing
- User registration and login/logout flows
- Authentication middleware
- Users, workspaces, workspace members, projects, repositories, and sessions in the domain layer
- Repository and service layers for users, workspaces, workspace members, projects, repositories, and sessions
- HTTP handlers and routes for health checks, authentication, workspaces, workspace members, and projects
- PostgreSQL configuration and database access
- SQL migrations for the current database schema
- Unit, service, repository, and HTTP tests across the backend
- Docker Compose development database setup

#### Current Focus

- Continue backend development from the existing foundation
- Keep authentication and workspace authorization correct as new features are added
- Add each implementation step to this log before moving to the next major capability

#### Change Entries

| Date | Change | Files or area | Tests and validation | Status |
|------|--------|---------------|----------------------|--------|
| 2026-09-17 | Added this working change log and continuation record | `README.md` | README reviewed; repository baseline checked | Complete |
| 2026-09-17 | Added comprehensive unit tests for Repository Service `Create`, including a recording repository fake | `apps/api/internal/service/repository_service_test.go` | `go test ./internal/service -run '^TestRepositoryServiceCreate'`; `DATABASE_URL=postgres://devflow:devflowpassword@localhost:5432/devflow?sslmode=disable go test ./...` | Complete |
| 2026-09-17 | Implemented and tested Repository Service `FindByID` with project-membership authorization and error propagation | `apps/api/internal/service/repository_service.go`, `apps/api/internal/service/repository_service_test.go` | `go test ./internal/service -run '^TestRepositoryServiceFindByID'`; `go test ./internal/service`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Implemented and tested Repository Service `ListByProjectID` with project-membership authorization and repository delegation | `apps/api/internal/service/repository_service.go`, `apps/api/internal/service/repository_service_test.go` | `go test ./internal/service -run '^TestRepositoryServiceListByProjectID'`; `go test ./internal/service`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Implemented and tested Repository Service `Delete` with project-membership authorization and error propagation | `apps/api/internal/service/repository_service.go`, `apps/api/internal/service/repository_service_test.go` | `go test ./internal/service -run '^TestRepositoryServiceDelete'`; `go test ./internal/service`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Added repository HTTP handlers and focused status/authentication tests for create, list, get, and delete | `apps/api/internal/http/repository_handler.go`, `apps/api/internal/http/repository_handler_test.go` | `go test ./internal/http -run '^TestRepositoryHandler'`; `go test ./internal/http`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Wired repository handlers into `router.go` and `main.go`, then added protected route coverage | `apps/api/internal/http/router.go`, `apps/api/internal/http/router_test.go`, `apps/api/cmd/server/main.go` | `go test ./internal/http -run '^TestRouter'`; `go test ./internal/http`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Verified the running API and documented the required database environment for the full test command | `README.md` | `DATABASE_URL=postgres://devflow:devflowpassword@localhost:5432/devflow?sslmode=disable go test ./...`; `curl http://localhost:8080/health` returned `{"service":"devflow-api","status":"ok"}` | Complete |
| 2026-09-17 | Added the GitHub connection domain contract, migration 8, status validation, and PostgreSQL uniqueness tests | `apps/api/internal/domain/github_connection.go`, `apps/api/internal/domain/github_connection_test.go`, `apps/api/internal/database/migrations/000008_create_github_connections.up.sql`, `apps/api/internal/database/migrations/000008_create_github_connections.down.sql`, `apps/api/internal/database/migrations/000008_create_github_connections_test.go` | `go test ./internal/domain`; focused migration constraint test; full `DATABASE_URL=postgres://devflow:devflowpassword@localhost:5432/devflow?sslmode=disable go test ./...` | Complete |
| 2026-09-17 | Added the GitHub connection repository interface, PostgreSQL implementation, and create/lookup/status-update integration tests | `apps/api/internal/repository/github_connection_repository.go`, `apps/api/internal/repository/github_connection_repository_test.go` | `DATABASE_URL=postgres://devflow:devflowpassword@localhost:5432/devflow?sslmode=disable go test ./internal/repository -run '^TestGitHubConnectionRepository'`; full `go test ./...` | Complete |
| 2026-09-17 | Added the GitHub connection service with member read access, owner/admin management authorization, duplicate protection, and status-transition tests | `apps/api/internal/service/github_connection_service.go`, `apps/api/internal/service/github_connection_service_test.go` | `go test ./internal/service -run '^TestGitHubConnectionService'`; `go test ./internal/service`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Added authenticated GitHub connection HTTP handlers and tests for create, lookup, status updates, authorization, validation, and transitions | `apps/api/internal/http/github_connection_handler.go`, `apps/api/internal/http/github_connection_handler_test.go` | `go test ./internal/http -run '^TestGitHubConnectionHandler'`; `go test ./internal/http`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Wired GitHub connection dependencies into `main.go` and protected workspace routes into `router.go`, with route-level tests | `apps/api/cmd/server/main.go`, `apps/api/internal/http/router.go`, `apps/api/internal/http/router_test.go` | `go test ./internal/http -run '^TestRouter'`; `go test ./internal/http`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Added typed GitHub App configuration validation and a mockable integration client interface without network calls | `apps/api/internal/config/config.go`, `apps/api/internal/config/config_test.go`, `apps/api/internal/integration/github/client.go`, `apps/api/internal/integration/github/client_test.go` | `go test ./internal/config ./internal/integration/github`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Implemented the GitHub App client with injectable HTTP transport, JWT authentication, installation-token exchange, repository decoding, and API error handling | `apps/api/internal/integration/github/client.go`, `apps/api/internal/integration/github/client_test.go` | `go test ./internal/integration/github`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Added the GitHub repository import service with connection lookup, GitHub client delegation, repository metadata mapping, duplicate skipping, and error propagation tests | `apps/api/internal/service/github_repository_import_service.go`, `apps/api/internal/service/github_repository_import_service_test.go` | `go test ./internal/service -run '^TestGitHubRepositoryImportService'`; `go test ./internal/service`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Exposed repository import through an authenticated endpoint and added a controlled unavailable-client fallback when GitHub App credentials are missing | `apps/api/internal/http/github_repository_import_handler.go`, `apps/api/internal/http/github_repository_import_handler_test.go`, `apps/api/internal/http/router.go`, `apps/api/internal/http/router_test.go`, `apps/api/cmd/server/main.go`, `apps/api/internal/integration/github/client.go` | `go test ./internal/http -run '^(TestGitHubRepositoryImportHandler|TestRouter)'`; `go test ./internal/http`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Added an in-process background import queue with bounded retries, cancellation handling, asynchronous `202` responses, and worker lifecycle wiring | `apps/api/internal/service/github_repository_import_job.go`, `apps/api/internal/service/github_repository_import_job_test.go`, `apps/api/internal/http/github_repository_import_handler.go`, `apps/api/cmd/server/main.go` | `go test ./internal/service`; `go test ./internal/http`; full `go test ./...` with Compose database URL | Complete |
| 2026-09-17 | Added durable import-job status persistence, migration 9, structured failure code/message fields, retry/success/failure metrics, and lifecycle integration tests | `apps/api/internal/domain/github_repository_import_job.go`, `apps/api/internal/database/migrations/000009_create_github_repository_import_jobs.*`, `apps/api/internal/repository/github_repository_import_job_repository.go`, `apps/api/internal/repository/github_repository_import_job_repository_test.go`, `apps/api/internal/service/github_repository_import_job.go` | Migration 9 applied; focused durable repository tests; full `go test ./...` with Compose database URL | Complete |

When development resumes, append a new row rather than replacing previous entries. This README is the source of truth for continuing work across sessions.

Next small step: expose a protected job-status lookup endpoint and add durable queue recovery for jobs left in `running` after process shutdown.

### Next Large Steps

The larger implementation sequence is:

1. Complete repository management
     - `ListByProjectID` and `Delete` service tests and implementations
     - HTTP handlers, routes, request validation, and authorization tests
     - Consistent API error responses

2. Stabilize the backend platform
     - Authentication and authorization review across every endpoint
     - Request-scoped context and structured error handling
     - Configuration, logging, health checks, and database migration discipline
     - CI test execution and integration-test database setup

3. Build GitHub integration
     - Secure GitHub connection and token storage model
     - Repository import and synchronization service
     - Issues and pull requests domain models and persistence
     - Rate-limit handling, retries, webhook validation, and sync observability

4. Add the frontend workspace
     - Authentication and workspace navigation
     - Projects, repositories, members, and activity views
     - Typed API client and consistent loading/error states

5. Add code intelligence
     - Repository ingestion and file metadata
     - Code and documentation indexing
     - Search and retrieval contracts
     - Background jobs and incremental re-indexing

6. Add the AI engineering assistant
     - Context assembly from repository, project, and documentation data
     - Provider abstraction and request safety limits
     - Conversation persistence, citations, and auditability

7. Harden for production
     - CI/CD deployment pipeline
     - Metrics, tracing, alerting, backups, and secret management
     - Security review, performance testing, and operational documentation

### GitHub Integration Design

The first GitHub integration will keep connection metadata separate from imported repository metadata.

#### Initial Decisions

- A GitHub connection belongs to a workspace, not to an individual repository.
- The preferred integration model is a GitHub App installation because it supports least-privilege repository access, installation scoping, and webhook delivery.
- Store the GitHub installation ID, account login, selected permissions, connection status, and timestamps in a dedicated connection model.
- Do not store long-lived GitHub access tokens in the `repositories` table or in logs. Generate short-lived installation tokens on the server using application credentials managed outside the database.
- Keep GitHub API behavior behind an integration client interface so services remain testable without network calls.
- Import metadata through the existing Repository Service so authorization and duplicate protection remain in one place.
- Run repository synchronization asynchronously after connection or import; HTTP requests should not wait for a complete repository sync.

#### Planned Flow

```text
Workspace user
     ↓
GitHub App installation / connection validation
     ↓
Persist workspace GitHub connection metadata
     ↓
List accessible GitHub repositories
     ↓
Repository Service Create
     ↓
Background synchronization and webhook updates
```

Next GitHub step: expose synchronization job status and add recovery for interrupted jobs.

---

## 🗺️ Development Roadmap

DevFlow is being developed through the following stages:

```text
Requirements
     ↓
Product Specification
     ↓
System Architecture
     ↓
Database Schema
     ↓
API Design
     ↓
Authentication & Security
     ↓
Backend
     ↓
Frontend
     ↓
GitHub Integration
     ↓
Code Intelligence / RAG
     ↓
AI Assistant
     ↓
CI/CD Intelligence
     ↓
Testing
     ↓
Docker
     ↓
CI/CD
     ↓
Observability
     ↓
Deployment
     ↓
Documentation
```

The roadmap is intentionally incremental so that each layer can be designed, implemented, tested, and validated before introducing the next major subsystem.

---

## 🎯 Project Goals

DevFlow AI is being built with production engineering principles in mind.

### Maintainability

Clear separation of concerns and modular architecture.

### Scalability

Architecture capable of evolving from a local development project into a multi-user engineering platform.

### Security

Authentication, authorization, data protection, and secure integrations are treated as core requirements.

### Reliability

Automated testing, database integrity, health checks, logging, and observability.

### Developer Experience

The platform should reduce context switching and make engineering information easier to discover.

### AI-Native Engineering

AI should work with the context of the actual project rather than functioning as a generic chatbot.

---

## 📌 Project Status

**Active development**

DevFlow AI is currently in the backend foundation stage. Core infrastructure and domain capabilities are being implemented incrementally before moving into the larger GitHub, frontend, code-intelligence, and AI systems.

---

## 📄 License

License information will be added as the project approaches its first public release.

---

## 👨‍💻 Author

**Christopher Okoh**

GitHub:

`https://github.com/kristopher1027`

---

> DevFlow AI is being built as an industry-grade software engineering project focused on combining developer productivity, project intelligence, GitHub integration, and AI-assisted software engineering into one platform.