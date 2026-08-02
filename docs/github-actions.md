# GitHub Actions

## 概要

この repo では GitHub Actions で CI / 環境変数チェック / CD を分けている。

```text
overview-ci
  Go test、Docker build、.env commit 防止

overview-env-check
  GitHub Secrets / Variables から .env を生成できるか確認

overview-deploy
  Tailscale 経由でホストへ SSH
  ホスト上の clone 済み repo を git pull 相当で更新
  GitHub の値から .env を生成
  Docker Compose で反映
```

## 重要な前提

GitHub が自動で「どのディレクトリに push するか」を選ぶわけではない。

deploy workflow が以下を使って、反映先を明示している。

```text
DEPLOY_HOST
DEPLOY_USER
DEPLOY_PATH
```

今回の方式では、デプロイ先ホストに事前に repo を clone しておく。

例:

```sh
sudo mkdir -p /srv/overview
sudo chown "$USER":"$USER" /srv/overview
git clone https://github.com/med-000/overview.git /srv/overview
cd /srv/overview
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

### overview-ci

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

### overview-env-check

実行タイミング:

```text
push: main
workflow_dispatch
```

内容:

- GitHub Secrets / Variables から `backend/.env` と `infra/.env` を生成
- 必須値が入っていなければ失敗
- 生成した `.env` で `docker compose config`

### overview-deploy

実行タイミング:

```text
push: main
workflow_dispatch
```

内容:

- GitHub-hosted runner が Tailscale に一時参加
- deploy host に SSH
- `DEPLOY_PATH` の clone 済み repo を `origin/main` に更新
- GitHub Secrets / Variables から `.env` を生成してホストへ配置
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
DEPLOY_HOST
DEPLOY_USER
DEPLOY_SSH_PRIVATE_KEY
DEPLOY_SSH_PORT
TS_OAUTH_CLIENT_ID
TS_OAUTH_SECRET
MATTERMOST_OVERVIEW_WEBHOOK
NOTION_API_KEY
NOITON_OVERVIEW_DB_KEY
```

`DEPLOY_SSH_PORT` は通常 `22`。省略しても workflow 側で `22` として扱う。

`DEPLOY_HOST` は Tailscale IP か MagicDNS hostname。

`TS_OAUTH_CLIENT_ID` と `TS_OAUTH_SECRET` は Tailscale 経由で SSH するために必要。
Tailscale OAuth client は `tag:github-actions` を使えるようにしておく。

任意の Secret:

```text
NOTION_WEBHOOK_VERIFICATION_TOKEN
```

Notion Webhook の初回 verification で取得した token。設定すると infra が `X-Notion-Signature` を検証する。

## Variables

必須:

```text
DEPLOY_PATH
```

任意:

```text
NOTION_SYNC_INTERVAL_SECONDS
```

初期値の例:

```text
DEPLOY_PATH=/srv/overview
NOTION_SYNC_INTERVAL_SECONDS=300
```

`DEPLOY_PATH` は必須。ホスト上の clone 済み repo の絶対パス。

`NOTION_SYNC_INTERVAL_SECONDS` は未設定でも `300` 秒として扱う。頻繁に変える可能性があるため、必要なら Variable にしてよい。

それ以外のポート、timeout、Notion API version などは細かすぎるので GitHub Variables には置かない。必要になったら `.env.example` と app config の default を見直す。

## 手動 deploy

接続とチェックだけ:

```text
Actions
-> overview-deploy
-> Run workflow
-> apply=check
```

反映まで行う:

```text
Actions
-> overview-deploy
-> Run workflow
-> apply=deploy
```

サービス再起動だけ:

```text
Actions
-> overview-deploy
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

private key は GitHub Secret `DEPLOY_SSH_PRIVATE_KEY` に入れる。

## 注意

`.env` は repo に commit しない。

本番 `.env` は GitHub Secrets / Variables から workflow が生成する。

main push 時は deploy が自動で走る。最初の動作確認では `workflow_dispatch` の `apply=check` から試す。
