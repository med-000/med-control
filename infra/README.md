# infra

外部サービス連携を担当する。現状は Notion から task を取得し、backend に同期する。

## 役割

- Notion DB の row/page を取得
- page 本文を Markdown に変換
- `shared/domain/task.Task` に整形
- backend に POST
- 毎日 01:00 台に、前日までの label が `schedule` かつ status が `done` ではない item を Notion status `done` にする
- `GET /notion/templates` で Notion data source template の一覧を返す
- `POST /notion/webhook` で Notion Webhook を受け、該当 page を即時同期
- `POST /tasks/quick` で backend から task 作成依頼を受け、Notion page を作る
- Mattermost Incoming Webhook 送信実装を `infra/mattermost` に保持

## 構成

```text
cmd/app                    Notion 同期 worker。5 分ごとの polling と Webhook 受信
cmd/raw                    Notion 生データ確認
cmd/tasks                  backend に渡す整形済み task 確認
internal/app/tasksync      同期 usecase
internal/control/httpapi   Notion Webhook handler
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
go run ./cmd/templates
```

## 環境変数

```env
NOTION_API_KEY=
NOTION_MED_CONTROL_DB_KEY=
BACKEND_TASKS_ENDPOINT=http://localhost:8080/tasks/import
NOTION_SYNC_INTERVAL_SECONDS=300
NOTION_WEBHOOK_ADDR=:8090
NOTION_WEBHOOK_VERIFICATION_TOKEN=
NOTION_API_BASE_URL=https://api.notion.com
NOTION_API_VERSION=2026-03-11
NOTION_PAGE_SIZE=100
HTTP_TIMEOUT_SECONDS=10
```

## Notion Templates

`/work` などで使う template ID を確認するために、infra は設定済み Notion DB の template 一覧を返せる。

```sh
curl http://localhost:8090/notion/templates
```

または server を起動せずに:

```sh
cd infra
go run ./cmd/templates
```

レスポンスの `templates[].id` を template ID として使う。

Notion GUI から template の link を copy した場合は、URL 内の page ID 部分を使う。

```text
https://app.notion.com/p/<workspace>-<template_id_without_hyphens>?v=...
                         ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
```

この例では以下が template ID。

```text
<template_id_without_hyphens>
```

API の `templates[].id` は hyphen 付きで返る。

```text
<template-id-with-hyphens>
```

hyphen を除いた値が GUI link の ID と一致すれば同じ template。
`.env` には hyphen 付きの API response を入れる。

## Notion Webhook

Notion connection の Webhooks で以下を登録する。

```text
Webhook URL: https://<public-host>/notion/webhook
Event types:
  page.created
  page.properties_updated
  page.content_updated
  data_source.content_updated
  data_source.schema_updated
```

初回登録時、Notion は `verification_token` を送る。infra のログに出る token を Notion の verify 画面に貼る。
本番では同じ token を `NOTION_WEBHOOK_VERIFICATION_TOKEN` に入れると `X-Notion-Signature` を検証する。

## Auto Completion

infra worker は毎日 01:00 台に前日までの未完了 schedule item を閉じる。

対象:

```text
label: schedule
date.start: 当日 00:00 より前
status: done 以外
```

処理:

```text
Notion status -> done
updated item -> backend に同期
```
