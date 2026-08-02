# GitHub Actions

## Workflows

- `overview-ci`: Go tests, Docker Compose config/build, and committed `.env` guard.
- `overview-env-check`: validates GitHub Secrets/Variables by generating `backend/.env` and `infra/.env`.
- `overview-deploy`: syncs source to the host, writes `.env` files from GitHub values, checks, and deploys with Docker Compose.

## GitHub Secrets

Set these in `Settings -> Secrets and variables -> Actions -> Environment secrets` for the `production` environment.

```text
DEPLOY_HOST
DEPLOY_USER
DEPLOY_SSH_PRIVATE_KEY
DEPLOY_SSH_PORT
MATTERMOST_OVERVIEW_WEBHOOK
NOTION_API_KEY
NOITON_OVERVIEW_DB_KEY
```

`DEPLOY_SSH_PORT` may be omitted if the host uses port `22`.

## GitHub Variables

Set these in the same `production` environment.

```text
DEPLOY_PATH
BACKEND_ADDR
TASK_NOTIFY_INTERVAL_SECONDS
HTTP_TIMEOUT_SECONDS
BACKEND_TASKS_ENDPOINT
NOTION_SYNC_INTERVAL_SECONDS
NOTION_API_BASE_URL
NOTION_API_VERSION
NOTION_PAGE_SIZE
```

Recommended initial values:

```text
DEPLOY_PATH=/srv/overview
BACKEND_ADDR=:8080
TASK_NOTIFY_INTERVAL_SECONDS=60
HTTP_TIMEOUT_SECONDS=10
BACKEND_TASKS_ENDPOINT=http://backend:8080/tasks/import
NOTION_SYNC_INTERVAL_SECONDS=300
NOTION_API_BASE_URL=https://api.notion.com
NOTION_API_VERSION=2026-03-11
NOTION_PAGE_SIZE=100
```

## Deploy

Manual check only:

```text
Actions -> overview-deploy -> Run workflow -> apply=check
```

Manual deploy:

```text
Actions -> overview-deploy -> Run workflow -> apply=deploy
```

Push to `main` runs deploy automatically after the host-side check passes.

The deploy host must have Docker, Docker Compose, `make`, and `rsync` available. If the host is only reachable over Tailscale, add a Tailscale connection step before `Prepare SSH`.
