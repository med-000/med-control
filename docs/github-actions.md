# GitHub Actions

## 概要

この repo では GitHub Actions で CI / 環境変数チェック / CD を分けている。

```text
med-control-ci
  Go test、Docker build、.env commit 防止

med-control-env-check
  GitHub Secrets から .env を生成できるか確認

med-control-deploy
  Tailscale 経由でホストへ SSH
  ホスト上の clone 済み repo を git pull 相当で更新
  GitHub Secrets の *_ENV_FILE から .env を生成
  Docker Compose で反映
```

## 重要な前提

GitHub が自動で「どのディレクトリに push するか」を選ぶわけではない。

deploy workflow が以下を使って、反映先を明示している。

```text
SSH_HOST
SSH_USER
DEPLOY_PATH
```

今回の方式では、デプロイ先ホストに事前に repo を clone しておく。

例:

```sh
sudo mkdir -p /srv/med-control
sudo chown "$USER":"$USER" /srv/med-control
git clone https://github.com/med-000/med-control.git /srv/med-control
cd /srv/med-control
git checkout main
```

Actions はその後、ホスト上で以下を実行する。

```sh
cd "$DEPLOY_PATH"
git fetch origin main
git reset --hard origin/main
```

つまり、反映先は GitHub 側が選ぶのではなく、`DEPLOY_PATH` で指定した clone 済みディレクトリ。

## Workflows

### med-control-ci

実行タイミング:

```text
push: main / develop
pull_request
workflow_dispatch
```

内容:

- `shared`, `infra`, `backend` の `go test ./...`
- dummy `.env` を使った `docker compose config`
- `docker compose build`
- `.env` が commit されていないか確認

### med-control-env-check

実行タイミング:

```text
push: main
workflow_dispatch
```

内容:

- GitHub Secrets の `BACKEND_ENV_FILE` と `INFRA_ENV_FILE` から `backend/.env` と `infra/.env` を生成
- 必須値が入っていなければ失敗
- 生成した `.env` で `docker compose config`

### med-control-deploy

実行タイミング:

```text
push: main
workflow_dispatch
```

内容:

- GitHub-hosted runner が Tailscale に一時参加
- deploy host に SSH
- `DEPLOY_PATH` の clone 済み repo を `origin/main` に更新
- GitHub Secrets の `BACKEND_ENV_FILE` と `INFRA_ENV_FILE` から `.env` を生成してホストへ配置
- ホスト上で `docker compose config`
- `docker compose up -d --build`

## GitHub Environment

`production` environment を使う。

場所:

```text
Repository
-> Settings
-> Secrets and variables
-> Actions
-> Environments
-> production
```

Environment はディレクトリ単位ではなく、deploy 先や実行環境単位で使う。

## 必要な Secrets

`production` environment の Secrets に入れる。

```text
SSH_HOST
SSH_USER
SSH_PRIVATE_KEY
SSH_PORT
TS_OAUTH_CLIENT_ID
TS_OAUTH_SECRET
BACKEND_ENV_FILE
INFRA_ENV_FILE
```

`BACKEND_ENV_FILE` には `backend/.env` の全文を入れる。

例:

```dotenv
BACKEND_ADDR=:8080
BACKEND_DB_PATH=/data/med-control.db
MATTERMOST_MED_CONTROL_WEBHOOK=https://...
MATTERMOST_COMMAND_TOKEN=...
MATTERMOST_REMIND_COMMAND_TOKEN=
MATTERMOST_QUICK_COMMAND_TOKEN=
INFRA_QUICK_TASK_ENDPOINT=http://infra:8080/tasks/quick
TASK_NOTIFY_INTERVAL_SECONDS=60
HTTP_TIMEOUT_SECONDS=10
```

`INFRA_ENV_FILE` には `infra/.env` の全文を入れる。

例:

```dotenv
NOTION_API_KEY=secret_...
NOTION_MED_CONTROL_DB_KEY=...
BACKEND_TASKS_ENDPOINT=http://backend:8080/tasks/import
NOTION_SYNC_INTERVAL_SECONDS=300
NOTION_WEBHOOK_ADDR=:8080
NOTION_WEBHOOK_VERIFICATION_TOKEN=
NOTION_API_BASE_URL=https://api.notion.com
NOTION_API_VERSION=2026-03-11
NOTION_PAGE_SIZE=100
HTTP_TIMEOUT_SECONDS=10
```

`SSH_PORT` は通常 `22`。省略しても workflow 側で `22` として扱う。

`SSH_HOST` は Tailscale IP か MagicDNS hostname。

`TS_OAUTH_CLIENT_ID` と `TS_OAUTH_SECRET` は Tailscale 経由で SSH するために必要。
Tailscale OAuth client は `tag:github-actions` を使えるようにしておく。

## Variables

必須:

```text
DEPLOY_PATH
```

初期値の例:

```text
DEPLOY_PATH=/srv/med-control
```

`DEPLOY_PATH` は必須。ホスト上の clone 済み repo の絶対パス。

app 設定は GitHub Variables に分けず、ディレクトリ名に対応した `BACKEND_ENV_FILE` / `INFRA_ENV_FILE` の Secret にまとめる。
`.env` の項目を増やす場合は `.env.example` と app config を更新し、対応する `*_ENV_FILE` の本文も更新する。

## 手動 deploy

接続とチェックだけ:

```text
Actions
-> med-control-deploy
-> Run workflow
-> apply=check
```

反映まで行う:

```text
Actions
-> med-control-deploy
-> Run workflow
-> apply=deploy
```

サービス再起動だけ:

```text
Actions
-> med-control-deploy
-> Run workflow
-> apply=restart
```

## ホスト側に必要なもの

deploy host には以下が必要。

```text
git
docker
docker compose
scp/ssh
```

GitHub Actions 用の SSH public key を deploy user の `~/.ssh/authorized_keys` に入れておく。

private key は GitHub Secret `SSH_PRIVATE_KEY` に入れる。

## 注意

`.env` は repo に commit しない。

本番 `.env` は GitHub Secrets の `BACKEND_ENV_FILE` / `INFRA_ENV_FILE` から workflow が生成する。

main push 時は deploy が自動で走る。最初の動作確認では `workflow_dispatch` の `apply=check` から試す。
