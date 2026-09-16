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
go test ./...
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
DATABASE_URL=postgres://devflow:devflow@localhost:5432/devflow?sslmode=disable
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