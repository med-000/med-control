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
  ホスト上の clone 済み repo を git pull --ff-only で更新
  GitHub Secrets の *_ENV_FILE から .env を生成
  Docker Compose で反映
```

## 重要な前提

GitHub が自動で「どのディレクトリに push するか」を選ぶわけではない。

deploy workflow は SSH 接続先だけを GitHub Secrets から受け取り、反映先 path は repo 側で固定している。

```text
SSH_HOST
SSH_USER
/srv/med-control
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
cd /srv/med-control
git pull --ff-only
```

つまり、反映先は GitHub 側が選ぶのではなく、この repo の deploy workflow / deploy script が決めた `/srv/med-control`。

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
- `/srv/med-control` の clone 済み repo を `git pull --ff-only` で更新
- GitHub Secrets の `BACKEND_ENV_FILE` と `INFRA_ENV_FILE` から `backend/.env` / `infra/.env` を生成してホストへ配置
- 変更 file から deploy action を決める
- `sudo -n /usr/local/bin/deploy-med-control <action>` を実行する

## GitHub Secrets

場所:

```text
Repository
-> Settings
-> Secrets and variables
-> Actions
-> Secrets
```

Environment は使わず、repository secrets に置く。

## 必要な Secrets

Repository Secrets に入れる。

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

この workflow では GitHub Variables は使わない。
deploy 先 path は `/srv/med-control` に固定し、app 設定は GitHub Variables に分けず、ディレクトリ名に対応した `BACKEND_ENV_FILE` / `INFRA_ENV_FILE` の Secret にまとめる。
`.env` の項目を増やす場合は `.env.example` と app config を更新し、対応する `*_ENV_FILE` の本文も更新する。

med4svc の `ENV_FILE` 1本方式とは違い、この repo は service directory ごとに `.env` を分ける。
workflow は remote 上で以下を atomically 置き換える。

```text
/srv/med-control/.env
/srv/med-control/backend/.env
/srv/med-control/infra/.env
```

root の `.env` は Docker Compose の `--env-file` 用で、通常は空 file。

## 手動 deploy

接続とチェックだけ:

```text
Actions
-> med-control-deploy
-> Run workflow
-> apply=check
```

再作成まで行う:

```text
Actions
-> med-control-deploy
-> Run workflow
-> apply=recreate
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

Docker Compose の永続データは deploy host の `/srv/med-control/data/` に置く。
`data/` は Git 追跡対象外。

app 内 Caddy は `med-control-caddy` という container name で `med2-gateway` network に参加する。
med2 側の `network/caddy/routes.yaml` では `upstream` に以下を指定する。

```yaml
upstream: med-control-caddy:8080
```

public path の振り分けは repo 内の `network/caddy/Caddyfile` が持つ。

```text
/mattermost/commands/* -> backend:8080
/notion/webhook        -> infra:8080
```

## deploy script

Git 管理している deploy script は `bin/deploy-med-control`。
これは server 上で直接その path を叩く実体ではなく、root owned script として `/usr/local/bin/deploy-med-control` に配置するための元ファイル。

server 側への配置:

```sh
sudo install -o root -g root -m 755 /srv/med-control/bin/deploy-med-control /usr/local/bin/deploy-med-control
```

deploy user に `docker group` や `sudo ALL` は渡さない。
Docker 操作は `/usr/local/bin/deploy-med-control` の固定 action だけに閉じ込める。

sudoers は `visudo` で作る。

```sh
sudo visudo -f /etc/sudoers.d/deploy-med-control
```

中身:

```sudoers
<deploy-user> ALL=(root) NOPASSWD: /usr/local/bin/deploy-med-control check
<deploy-user> ALL=(root) NOPASSWD: /usr/local/bin/deploy-med-control pull
<deploy-user> ALL=(root) NOPASSWD: /usr/local/bin/deploy-med-control reload
<deploy-user> ALL=(root) NOPASSWD: /usr/local/bin/deploy-med-control restart
<deploy-user> ALL=(root) NOPASSWD: /usr/local/bin/deploy-med-control recreate
<deploy-user> ALL=(root) NOPASSWD: /usr/local/bin/deploy-med-control ps
```

workflow は host 上の repo と `.env` を更新した後、この script を `sudo -n` で呼ぶ。
`sudo -n` は password prompt を出さないため、sudoers 設定が足りない場合は CI/CD 上で即失敗する。

## 注意

`.env` は repo に commit しない。

本番 `.env` は GitHub Secrets の `BACKEND_ENV_FILE` / `INFRA_ENV_FILE` から workflow が生成する。

main push 時は deploy が自動で走る。最初の動作確認では `workflow_dispatch` の `apply=check` から試す。
