#!/usr/bin/env sh
set -eu

write_env_file() {
  target="$1"
  var_name="$2"
  value="$(printenv "$var_name" || true)"

  if [ -z "$value" ]; then
    echo "missing required environment variable: $var_name" >&2
    exit 1
  fi

  mkdir -p "$(dirname "$target")"
  printf '%s\n' "$value" > "$target"
}

write_env_file backend/.env BACKEND_ENV_FILE
write_env_file infra/.env INFRA_ENV_FILE

echo "wrote backend/.env and infra/.env"
