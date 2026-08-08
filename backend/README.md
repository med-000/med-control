# backend

アプリの中心。`infra` から受け取った task を保持し、READ API と通知判定を提供する。

## 役割

- `POST /tasks/import` で task を取り込む
- `GET /tasks` で保持中の task を返す
- `POST /tasks/notify-due` で通知時刻を過ぎた task を Mattermost に通知する
- `POST /mattermost/commands/remind` で `/remind <No> <minutes>` を受ける
- `POST /mattermost/commands/create` で `/create <title> <date> <notification>` を受け、infra 経由で Notion task を作る
- `POST /mattermost/commands/quick` で `/quick <title> <date> <notification>` を受け、priority High の Notion task を作る
- `POST /mattermost/commands/work` で `/work <start|end|todo> <start_mmdd> <end_mmdd>` を受け、Notion template 付きの勤務 item を作る
- `TASK_NOTIFY_INTERVAL_SECONDS` を設定すると定期的に通知判定する

task の local snapshot、`/remind` の一時通知時刻、通知済み履歴は SQLite に保存する。
Notion を正本とし、SQLite は med-control 側の cache / 制御状態 / 履歴を持つ。

## 構成

```text
cmd/app                       起動口
internal/app/task             task usecase
internal/control/httpapi      HTTP handler
internal/control/sqlite       SQLite repository
internal/control/memory       test/local in-memory repository
internal/config               env config
```

Mattermost webhook の具体実装は `infra/mattermost` に置き、backend はそれを注入して使う。

`internal/app/task` が usecase と port を持つ。
`internal/control/*` は HTTP / SQLite / 外部 service への adapter とし、filter 判定や通知 key 生成のような domain rule は `shared/domain/task` に置く。

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
BACKEND_DB_PATH=/data/med-control.db
MATTERMOST_MED_CONTROL_WEBHOOK=
MATTERMOST_COMMAND_TOKEN=
MATTERMOST_REMIND_COMMAND_TOKEN=
MATTERMOST_QUICK_COMMAND_TOKEN=
MATTERMOST_CREATE_COMMAND_TOKEN=
MATTERMOST_WORK_COMMAND_TOKEN=
INFRA_QUICK_TASK_ENDPOINT=http://localhost:8090/tasks/quick
NOTION_WORK_TEMPLATE_ID=
TASK_NOTIFY_INTERVAL_SECONDS=60
HTTP_TIMEOUT_SECONDS=10
```

`BACKEND_DB_PATH` には通知済み履歴を保存する SQLite file を指定する。
Docker Compose では repo 配下の `data/backend/` を container の `/data` に mount し、`data/backend/med-control.db` に保存する。
`data/` は Git 追跡対象外。
table 設計は `docs/database.md` を参照。

`MATTERMOST_COMMAND_TOKEN` は `/remind`、`/create`、`/quick`、`/work` 共通 token として使える。
Mattermost 側で command ごとに token が別になる場合は、`MATTERMOST_REMIND_COMMAND_TOKEN`、`MATTERMOST_CREATE_COMMAND_TOKEN`、`MATTERMOST_QUICK_COMMAND_TOKEN`、`MATTERMOST_WORK_COMMAND_TOKEN` を使う。

`NOTION_WORK_TEMPLATE_ID` には `ollo勤務` template の ID を入れる。

## Mattermost Slash Commands

Mattermost 側で4つ登録する。
slash command の応答は Mattermost mobile でも見えるように `in_channel` で返す。

```text
Trigger Word: remind
Request URL: https://<backend-public-host>/mattermost/commands/remind
Method: POST
Usage: /remind <No> <minutes>
```

```text
Trigger Word: create
Request URL: https://<backend-public-host>/mattermost/commands/create
Method: POST
Usage: /create <title> <date> <notification>
```

```text
Trigger Word: quick
Request URL: https://<backend-public-host>/mattermost/commands/quick
Method: POST
Usage: /quick <title> <date> <notification>
```

```text
Trigger Word: work
Request URL: https://<backend-public-host>/mattermost/commands/work
Method: POST
Usage: /work <start|end|todo> <start_mmdd> <end_mmdd>
```

`date` は `2026-08-10`、`notification` は `09:30` または `2026-08-10T09:30` を受け付ける。
`notification` が時刻だけの場合は `date` と同じ日として扱う。

`/work` の `mmdd` は現在年として扱う。`start` は Notion status `inprogress`、`end` は `done`、`todo` は `todo` に変換する。
