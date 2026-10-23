# DevFlow AI staging deployment

This Compose stack runs the React frontend, Go API, PostgreSQL, and versioned database migrations. PostgreSQL is not published to the host; only the frontend proxy port is exposed.

## Requirements

- Docker Engine and Docker Compose v2
- A strong staging-only PostgreSQL password
- GitHub App credentials to enable GitHub repository import/sync
- An AI provider key if you want AI answers/explanations

## Configure secrets

Create a deployment-only `.env.staging` file in the repository root. Do not commit it.

```dotenv
POSTGRES_PASSWORD=replace-with-a-long-random-alphanumeric-secret
STAGING_PORT=8081
GITHUB_APP_ID=
GITHUB_APP_PRIVATE_KEY=
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GEMINI_API_KEY=
```

Restrict access to this file:

```sh
chmod 600 .env.staging
```

Use a long random alphanumeric database password so it is safe inside the PostgreSQL URL. Supply only the AI provider key actually used by your configured provider. Keep GitHub App private keys and AI keys out of source control and logs.

For `GITHUB_APP_PRIVATE_KEY`, use the private-key format supported by the app's GitHub integration. If the key is stored as a single-line escaped value in your secret manager, verify the integration normalizes escaped newlines before relying on it.

## Start staging

Run these commands from the repository root:

```sh
docker compose --env-file .env.staging -f docker-compose.staging.yml config --quiet
docker compose --env-file .env.staging -f docker-compose.staging.yml build
docker compose --env-file .env.staging -f docker-compose.staging.yml up -d
docker compose --env-file .env.staging -f docker-compose.staging.yml ps
curl --fail http://localhost:8081/health
```

Open `http://localhost:8081` for the UI. The frontend's `/api/*` requests are proxied to the API with the `/api` prefix removed, matching the Vite development proxy. The secure session cookie setting is enabled; use HTTPS at the public staging ingress before exposing the stack over the internet.

## Logs and migrations

```sh
docker compose --env-file .env.staging -f docker-compose.staging.yml logs --tail=200 api web migrate
docker compose --env-file .env.staging -f docker-compose.staging.yml logs --tail=100 postgres
```

Migrations run before the API starts. If a migration fails, the API remains stopped until the migration issue is resolved.

## Database backup and restore

Before upgrading a staging database, create a logical backup:

```sh
umask 077
docker compose --env-file .env.staging -f docker-compose.staging.yml exec -T postgres sh -c 'pg_dump -U devflow -d devflow' > "devflow-staging-$(date +%Y%m%d-%H%M%S).sql"
```

Store backups outside the deployment host in encrypted storage and test restoration regularly. Restore only into an intentionally selected target database; never pipe a backup into a production database without checking the target first.

## Staging limitations

This file gives the app a reproducible containerized staging stack, but it does not create a public hosting account, domain, TLS certificate, managed database, external backup target, or monitoring service. For an internet-facing staging deployment, put HTTPS ingress in front of the web container, set the deployment secrets in the host's secret manager, and configure off-host backups and uptime/error monitoring before sharing the URL.
