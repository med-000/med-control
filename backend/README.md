# backend

アプリの中心。`infra` から受け取った task を保持し、READ API と通知判定を提供する。

## 役割

- `POST /tasks/import` で task を取り込む
- `GET /tasks` で保持中の task を返す
- `POST /tasks/notify-due` で通知時刻を過ぎた task を Mattermost に通知する
- `TASK_NOTIFY_INTERVAL_SECONDS` を設定すると定期的に通知判定する

現状の保存先は in-memory。再起動すると保持 task と通知済み状態は消える。

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
MATTERMOST_OVERVIEW_WEBHOOK=
TASK_NOTIFY_INTERVAL_SECONDS=60
HTTP_TIMEOUT_SECONDS=10
```
