# infra

外部サービス連携を担当する。現状は Notion から task を取得し、backend に同期する。

## 役割

- Notion DB の row/page を取得
- page 本文を Markdown に変換
- `shared/domain/task.Task` に整形
- backend に POST
- Mattermost Incoming Webhook 送信実装を `infra/mattermost` に保持

## 構成

```text
cmd/app                    Notion 同期 worker。5 分ごとに同期
cmd/raw                    Notion 生データ確認
cmd/tasks                  backend に渡す整形済み task 確認
internal/app/tasksync      同期 usecase
internal/control/notion    Notion API client / mapper / schema
internal/control/backend   backend への送信
mattermost                 Mattermost webhook 実装
```

Notion DB のカラム名は `internal/control/notion/schema.go` にまとめる。

## 実行

```sh
make infra
make infra-raw
make infra-tasks
```

## 環境変数

```env
NOTION_API_KEY=
NOITON_OVERVIEW_DB_KEY=
BACKEND_TASKS_ENDPOINT=http://localhost:8080/tasks/import
NOTION_SYNC_INTERVAL_SECONDS=300
NOTION_API_BASE_URL=https://api.notion.com
NOTION_API_VERSION=2026-03-11
NOTION_PAGE_SIZE=100
HTTP_TIMEOUT_SECONDS=10
```
