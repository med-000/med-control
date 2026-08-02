# overview

Notion と Mattermost を使ったカレンダー兼タスク管理の自動化・通知アプリ。

## 構成

```text
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

`infra` service は 5 分ごとに Notion を読み取り、`backend` に task を同期する。host から backend を見る場合は `http://localhost:8085` を使う。

## CI/CD

GitHub Actions は `.github/workflows` に定義している。

- `overview-ci`: Go / Docker / `.env` commit 防止の CI
- `overview-env-check`: GitHub Secrets/Variables から `.env` を生成できるか確認
- `overview-deploy`: ホストへ rsync し、`.env` を生成して Docker Compose で deploy

詳細は `docs/github-actions.md` を参照。
