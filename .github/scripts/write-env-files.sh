#!/usr/bin/env sh
set -eu

require_env() {
  name="$1"
  value="$(printenv "$name" || true)"
  if [ -z "$value" ]; then
    echo "missing required environment variable: $name" >&2
    exit 1
  fi
}

optional_env() {
  name="$1"
  fallback="$2"
  value="$(printenv "$name" || true)"
  if [ -n "$value" ]; then
    printf '%s' "$value"
    return
  fi
  printf '%s' "$fallback"
}

require_env MATTERMOST_OVERVIEW_WEBHOOK
require_env NOTION_API_KEY
require_env NOITON_OVERVIEW_DB_KEY

mkdir -p backend infra

cat > backend/.env <<EOF
BACKEND_ADDR=$(optional_env BACKEND_ADDR ':8080')
MATTERMOST_OVERVIEW_WEBHOOK=${MATTERMOST_OVERVIEW_WEBHOOK}
TASK_NOTIFY_INTERVAL_SECONDS=$(optional_env TASK_NOTIFY_INTERVAL_SECONDS '60')
HTTP_TIMEOUT_SECONDS=$(optional_env HTTP_TIMEOUT_SECONDS '10')
EOF

cat > infra/.env <<EOF
NOTION_API_KEY=${NOTION_API_KEY}
NOITON_OVERVIEW_DB_KEY=${NOITON_OVERVIEW_DB_KEY}
BACKEND_TASKS_ENDPOINT=$(optional_env BACKEND_TASKS_ENDPOINT 'http://backend:8080/tasks/import')
NOTION_SYNC_INTERVAL_SECONDS=$(optional_env NOTION_SYNC_INTERVAL_SECONDS '300')
NOTION_WEBHOOK_ADDR=$(optional_env NOTION_WEBHOOK_ADDR ':8080')
NOTION_WEBHOOK_VERIFICATION_TOKEN=$(optional_env NOTION_WEBHOOK_VERIFICATION_TOKEN '')
NOTION_API_BASE_URL=$(optional_env NOTION_API_BASE_URL 'https://api.notion.com')
NOTION_API_VERSION=$(optional_env NOTION_API_VERSION '2026-03-11')
NOTION_PAGE_SIZE=$(optional_env NOTION_PAGE_SIZE '100')
HTTP_TIMEOUT_SECONDS=$(optional_env HTTP_TIMEOUT_SECONDS '10')
EOF

echo "wrote backend/.env and infra/.env"
