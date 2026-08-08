# med-control

Notion と Mattermost を使ったカレンダー兼タスク管理の自動化・通知アプリ。

## 構成

```text
network/  app 内 gateway。外部入口を backend / infra に振り分ける
backend/   アプリの中心。タスク保持、READ API、通知判定を担当
infra/     外部サービス連携。Notion 取得、Mattermost webhook 実装を担当
shared/    backend と infra で共有する domain 型
frontend/  今後の frontend 置き場
docs/      設計・運用メモ
```

## よく使うコマンド

```sh
make backend      # backend を起動
make infra        # Notion 同期 worker を起動
make infra-raw    # Notion の生データを見る
make infra-tasks  # backend に渡す整形済み task を見る
make test         # Go test
make up           # Docker Compose 起動
make down         # Docker Compose 停止
```

## Docker

```sh
docker compose up --build
```

`infra` service は 5 分ごとに Notion を読み取り、Notion Webhook 受信時は該当 page を即時に取り直して `backend` に task を同期する。host から直接確認する場合は、local 起動用の `.env.example` を使って `make backend` と `make infra` を別 port で起動する。

Compose では app 内 Caddy だけを `med2-gateway` network に参加させる。med2 側の Caddy は `med-control-caddy:8080` に reverse proxy し、app 内 Caddy が `/mattermost/commands/*` を `backend`、`/notion/webhook` を `infra` に振り分ける。

Docker Compose の永続データは repo 配下の `data/` に置く。`data/` は Git 追跡対象外。

## CI/CD

GitHub Actions は `.github/workflows` に定義している。

- `med-control-ci`: Go / Docker / `.env` commit 防止の CI
- `med-control-env-check`: GitHub Secrets から `.env` を生成できるか確認
- `med-control-deploy`: ホスト上の repo を更新し、`.env` を生成して deploy script 経由で反映

詳細は `docs/github-actions.md` を参照。
