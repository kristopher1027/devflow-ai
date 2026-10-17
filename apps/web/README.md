# DevFlow web

A small React + TypeScript app for the DevFlow API: log in, create a workspace,
connect GitHub, import and sync a repository, then ask questions about its code.

## Run it

1. Start the Go API (it listens on port 8080).
2. In this folder:

```bash
npm install
npm run dev
```

3. Open http://localhost:5173 (in a Codespace, open the forwarded port 5173
   from the Ports tab).

The browser only talks to `/api/...` on the same address as the page. Vite
forwards those requests to the Go API (default `http://localhost:8080`), so
there is no CORS setup and the session cookie works. To use a different API:

```bash
DEVFLOW_API=http://localhost:9000 npm run dev
```

## Scripts

| Command             | What it does                                  |
| ------------------- | --------------------------------------------- |
| `npm run dev`       | Dev server with hot reload                    |
| `npm test`          | Unit tests (API client, routing, formatting)  |
| `npm run typecheck` | TypeScript check                              |
| `npm run build`     | Type-check and build to `dist/`               |
| `npm run preview`   | Serve the built app (with the same proxy)     |

## Pages

`#/` workspaces, `#/w/<id>` workspace (GitHub connection, projects),
`#/p/<id>` project (repositories, import, sync), `#/r/<id>` repository
(ask a question, explain the repository).
