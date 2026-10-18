# DevFlow AI

![CI](https://github.com/kristopher1027/devflow-ai/actions/workflows/ci.yml/badge.svg)

**Connect a GitHub repository and ask questions about its code.** DevFlow imports your repositories through a GitHub App, stores a snapshot of their files, finds the code relevant to a question, and has an AI model answer with links to the exact lines it used.

> Status: early stage, in active development. The backend, web app and AI answers work end to end locally. Deployment is the next milestone.

## What it does

1. **Create a workspace and a project**, then connect your GitHub account by installing a GitHub App with read-only access to repository contents.
2. **Import** the repositories the app can see.
3. **Sync** a repository: DevFlow records a snapshot of the default branch at its latest commit, downloads its source and documentation files, and splits them into searchable chunks.
4. **Ask** a question in plain English. DevFlow searches the chunks, builds a bounded context from the best matches, and asks the AI model to answer using only that code. The answer comes with clickable sources pinned to the synced commit.
5. **Explain** a repository in one click: what it does, its tech stack, structure and how to run it.

## Architecture

```mermaid
flowchart LR
    Browser["React + TypeScript app"] -->|"/api (same origin)"| API["Go API"]
    API --> DB[("PostgreSQL")]
    API -->|"GitHub App JWT"| GitHub["GitHub API"]
    API -->|"question + retrieved code"| AI["Gemini or Claude"]
```

The API is layered: HTTP handlers call services, services hold the rules, and repositories talk to PostgreSQL. Each layer depends on interfaces, so every piece is tested on its own with fakes.

```text
apps/
  api/                     Go API
    cmd/server/            wiring and start-up
    internal/
      auth/ config/ database/ domain/
      http/                handlers, middleware, routing
      service/             business rules, workers, retrieval, AI orchestration
      repository/          PostgreSQL access
      integration/         github, anthropic, gemini, aiprovider
  web/                     React + TypeScript + Vite front end
docs/                      archived plan and change log
docker-compose.yml         local PostgreSQL
```

### How a question gets answered

`Import` finds repositories, then `Sync` creates a snapshot and a background worker ingests the files and chunks them. At question time the keywords are extracted, the chunks matching the most keywords rank first, the retrieval builder fits them into a size budget, and the packager turns them into a prompt that labels the code as untrusted data. The model's answer is returned with the source files and line ranges.

## Engineering highlights

- **Durable background jobs.** Imports and syncs run on a worker queue with retries and status polling. Jobs interrupted by a restart are marked failed on start-up instead of hanging forever.
- **Security by default.** Session cookies are HttpOnly and only a hash of the session token is stored. Workspace roles are checked in the service layer. Secrets come from the environment and are never logged. Provider errors report only a status code.
- **Prompt-injection awareness.** Repository files are treated as untrusted input: the model is told never to follow instructions found inside them, and file content cannot close the prompt's wrapper tags.
- **Provider-agnostic AI.** Answers go through a small interface with Gemini (free tier) and Claude implementations, chosen by environment variable.
- **Tested.** Over 400 tests cover the API, including PostgreSQL integration tests, plus unit tests for the web client. CI builds the API, applies the migrations and runs the Go tests on every push.

## Run it locally

You need Docker, Go (see `apps/api/go.mod`), Node 22 and the [`migrate`](https://github.com/golang-migrate/migrate) CLI.

**1. Database**

```bash
docker compose up -d
migrate -path apps/api/internal/database/migrations \
  -database 'postgres://devflow:devflowpassword@localhost:5432/devflow?sslmode=disable' up
```

**2. A GitHub App** (one time)

Create a GitHub App at https://github.com/settings/apps/new with **Contents: Read-only**, no webhook, installable on your account only. Generate a private key, install the app on a repository, and note the App ID and the installation ID (the number in `github.com/settings/installations/<number>`).

**3. Configure and start the API**

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string |
| `GITHUB_APP_ID` | your GitHub App's ID |
| `GITHUB_APP_PRIVATE_KEY` | the App's private key (the PEM text) |
| `GEMINI_API_KEY` or `ANTHROPIC_API_KEY` | enables AI answers (optional) |
| `AI_PROVIDER` | `gemini` or `anthropic`, to force one (optional) |
| `GEMINI_MODEL` / `ANTHROPIC_MODEL` | override the default model (optional) |
| `PORT`, `DEVFLOW_COOKIE_SECURE` | listen port (default 8080) and the cookie Secure flag |

```bash
cd apps/api
export DATABASE_URL='postgres://devflow:devflowpassword@localhost:5432/devflow?sslmode=disable'
go run ./cmd/server
```

Without an AI key the server still starts, and the Ask and Explain buttons say AI is not configured.

**4. Start the web app**

```bash
cd apps/web
npm install
npm run dev
```

Open http://localhost:5173, create an account, create a workspace, connect GitHub with your installation ID, then create a project and press **Import from GitHub**. The browser only talks to `/api/...` on its own address, and Vite forwards that to the Go API, so there is no CORS setup.

**5. Optional: check the whole flow from the command line**

```bash
INSTALLATION_ID=<number> ACCOUNT_LOGIN=<github-username> bash devflow-live-test.sh
```

## API overview

All routes except register and login need the session cookie.

| Method and path | Purpose |
| --- | --- |
| `POST /auth/register`, `/auth/login`, `/auth/logout` | accounts and sessions |
| `GET`, `POST /workspaces` | list and create workspaces |
| `GET`, `POST /workspaces/{id}/projects` | projects in a workspace |
| `GET`, `POST`, `PATCH /workspaces/{id}/github` | the workspace's GitHub connection |
| `POST /projects/{id}/repositories/import` | import repositories (background job) |
| `GET /projects/{id}/repositories` | list repositories |
| `POST /repositories/{id}/sync` | sync a repository (background job) |
| `GET /github/repository-import-jobs/{id}`, `/repositories/sync-jobs/{id}` | job status |
| `POST /repositories/{id}/ask` | answer a question, with sources |
| `POST /repositories/{id}/explain` | explain the repository |

## Tests

```bash
cd apps/api && go test ./...      # repository tests need DATABASE_URL and a migrated database
cd apps/web && npm test           # API client, routing and error-message tests
```

## Roadmap

- Deploy a public demo (API, database and web app).
- Semantic search with embeddings (pgvector) in addition to keyword matching.
- Replace or remove a workspace's GitHub connection from the UI.
- Webhooks to re-sync automatically on push.
- Pull request and issue context, and richer repository analysis.

The original, much larger plan is kept in [docs/ROADMAP_AND_LOG.md](docs/ROADMAP_AND_LOG.md).

## License

MIT. See [LICENSE](LICENSE).

## Author

**Christopher Okoh**: [github.com/kristopher1027](https://github.com/kristopher1027)
