#!/usr/bin/env bash
# Live end-to-end test for DevFlow. The server must already be running.
#
#   INSTALLATION_ID=12345678 ACCOUNT_LOGIN=your-github-username bash devflow-live-test.sh
#
# It creates a throw-away user, workspace and project, connects your GitHub
# App installation, imports your repositories, syncs one, then (if the server
# has an Anthropic key) asks for an explanation. It never prints passwords.
set -euo pipefail

BASE="${BASE:-http://localhost:8080}"
: "${INSTALLATION_ID:?set INSTALLATION_ID (the number in your GitHub App installation URL)}"
: "${ACCOUNT_LOGIN:?set ACCOUNT_LOGIN (your GitHub username)}"

JAR="$(mktemp)"
BODY_FILE="$(mktemp)"
trap 'rm -f "$JAR" "$BODY_FILE"' EXIT

STAMP="$(date +%s)"
EMAIL="livetest-${STAMP}@example.com"
PASSWORD="Live-Test-${STAMP}-Pw!"

STATUS=""
BODY=""

step() { printf '\n== %s\n' "$*"; }

call() { # call METHOD PATH [JSON_BODY]
  local method="$1" path="$2" data="${3:-}"
  local args=(-sS -o "$BODY_FILE" -w '%{http_code}' -X "$method" -b "$JAR" -c "$JAR" "$BASE$path")
  if [ -n "$data" ]; then
    args+=(-H 'Content-Type: application/json' -d "$data")
  fi
  STATUS="$(curl "${args[@]}")"
  BODY="$(cat "$BODY_FILE")"
}

expect() { # expect STATUS...
  local ok
  for ok in "$@"; do
    if [ "$STATUS" = "$ok" ]; then return 0; fi
  done
  echo "FAILED: got HTTP $STATUS"
  echo "$BODY"
  exit 1
}

# field PYTHON_EXPR  -> evaluates against the last response body as `d`
field() {
  python3 -c "import sys,json; d=json.load(sys.stdin); print($1)" <<<"$BODY"
}

wait_for_job() { # wait_for_job PATH
  local i state
  for i in $(seq 1 60); do
    call GET "$1"
    expect 200
    state="$(field "d.get('Status') or d.get('status')")"
    case "$state" in
      succeeded) echo "job succeeded"; return 0 ;;
      failed)
        echo "job FAILED:"
        field "d.get('FailureCode') or d.get('failure_code')"
        field "d.get('FailureMessage') or d.get('failure_message')"
        exit 1
        ;;
    esac
    sleep 2
  done
  echo "FAILED: job still '$state' after 2 minutes"
  exit 1
}

step "health"
call GET /health
expect 200
echo "server is up"

step "register + login ($EMAIL)"
call POST /auth/register "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
expect 200 201
call POST /auth/login "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
expect 200
echo "logged in"

step "create workspace"
call POST /workspaces '{"name":"live-test"}'
expect 200 201
WS_ID="$(field "d['ID']")"
echo "workspace $WS_ID"

step "create project"
call POST "/workspaces/$WS_ID/projects" '{"name":"live-test-project"}'
expect 200 201
PROJECT_ID="$(field "d['ID']")"
echo "project $PROJECT_ID"

step "connect GitHub App installation"
call POST "/workspaces/$WS_ID/github" \
  "{\"installation_id\":\"$INSTALLATION_ID\",\"account_login\":\"$ACCOUNT_LOGIN\"}"
expect 200 201
CONNECTION_ID="$(field "d.get('id') or d.get('ID')")"
call PATCH "/workspaces/$WS_ID/github" "{\"id\":\"$CONNECTION_ID\",\"status\":\"active\"}"
expect 200 204
echo "connection $CONNECTION_ID is active"

step "import repositories from GitHub"
call POST "/projects/$PROJECT_ID/repositories/import"
expect 202
IMPORT_JOB="$(field "d['job_id']")"
wait_for_job "/github/repository-import-jobs/$IMPORT_JOB"

step "list imported repositories"
call GET "/projects/$PROJECT_ID/repositories"
expect 200
REPO_ID="$(field "next((r['ID'] for r in d if r.get('FullName','').endswith('/devflow-ai')), d[0]['ID'] if d else '')")"
if [ -z "$REPO_ID" ]; then
  echo "FAILED: no repositories were imported. Is the GitHub App installed on at least one repository?"
  exit 1
fi
echo "repositories imported: $(field "', '.join(r['FullName'] for r in d)")"
echo "using repository $REPO_ID"

step "sync repository (snapshot + files + chunks)"
call POST "/repositories/$REPO_ID/sync"
expect 202
SYNC_JOB="$(field "d['ID']")"
wait_for_job "/repositories/sync-jobs/$SYNC_JOB"

step "explain repository (needs ANTHROPIC_API_KEY on the server)"
call POST "/repositories/$REPO_ID/explain"
case "$STATUS" in
  200)
    echo "AI explanation:"
    field "d['explanation']"
    ;;
  503)
    echo "AI is not configured on the server (expected if you have no API key yet)."
    ;;
  *)
    echo "unexpected result: HTTP $STATUS"
    echo "$BODY"
    ;;
esac

echo
echo "DONE. Repository id for database checks: $REPO_ID"
