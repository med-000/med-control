# backend

アプリの中心。`infra` から受け取った task を保持し、READ API と通知判定を提供する。

## 役割

- `POST /tasks/import` で task を取り込む
- `GET /tasks` で保持中の task を返す
- `POST /tasks/notify-due` で通知時刻を過ぎた task を Mattermost に通知する
- `POST /mattermost/commands/remind` で `/remind <No> <minutes>` を受ける
- `POST /mattermost/commands/quick` で `/quick <title>` を受け、infra 経由で Notion task を作る
- `TASK_NOTIFY_INTERVAL_SECONDS` を設定すると定期的に通知判定する

現状の task 保存先は in-memory。再起動すると保持 task と `/remind` の一時通知時刻は消える。
通知済み履歴は SQLite に保存するため、再起動後も同じ通知時刻の二重通知を避けられる。

## 構成

```text
cmd/app                       起動口
internal/app/task             task usecase
internal/control/httpapi      HTTP handler
internal/control/memory       in-memory repository
internal/config               env config
```

Mattermost webhook の具体実装は `infra/mattermost` に置き、backend はそれを注入して使う。

## 実行

```sh
make backend
```

または:

```sh
go run ./cmd/app
```

## 環境変数

```env
BACKEND_ADDR=:8080
BACKEND_DB_PATH=/data/overview.db
MATTERMOST_OVERVIEW_WEBHOOK=
MATTERMOST_COMMAND_TOKEN=
MATTERMOST_REMIND_COMMAND_TOKEN=
MATTERMOST_QUICK_COMMAND_TOKEN=
INFRA_QUICK_TASK_ENDPOINT=http://localhost:8090/tasks/quick
TASK_NOTIFY_INTERVAL_SECONDS=60
HTTP_TIMEOUT_SECONDS=10
```

`BACKEND_DB_PATH` には通知済み履歴を保存する SQLite file を指定する。
Docker Compose では `/data/overview.db` を named volume に保存する。

`MATTERMOST_COMMAND_TOKEN` は `/remind` と `/quick` 共通 token として使える。
Mattermost 側で command ごとに token が別になる場合は、`MATTERMOST_REMIND_COMMAND_TOKEN` と `MATTERMOST_QUICK_COMMAND_TOKEN` を使う。

## Mattermost Slash Commands

Mattermost 側で2つ登録する。
slash command の応答は Mattermost mobile でも見えるように `in_channel` で返す。

```text
Trigger Word: remind
Request URL: https://<backend-public-host>/mattermost/commands/remind
Method: POST
Usage: /remind <No> <minutes>
```

```text
Trigger Word: quick
Request URL: https://<backend-public-host>/mattermost/commands/quick
Method: POST
Usage: /quick <title>
```
