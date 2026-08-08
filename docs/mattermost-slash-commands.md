# Mattermost Slash Commands

Mattermost 側で slash command を作り、Request URL を backend の `/mattermost/commands/*` に向ける。

## Token

Mattermost は slash command ごとに token を発行する。
backend は request form の `token` と `.env` の値を照合する。

```dotenv
MATTERMOST_COMMAND_TOKEN=
MATTERMOST_REMIND_COMMAND_TOKEN=
MATTERMOST_CREATE_COMMAND_TOKEN=
MATTERMOST_QUICK_COMMAND_TOKEN=
MATTERMOST_WORK_COMMAND_TOKEN=
```

`MATTERMOST_COMMAND_TOKEN` は fallback 用の共通 token。
command ごとに token を分ける場合は個別の `MATTERMOST_*_COMMAND_TOKEN` を使う。

## Commands

### remind

```text
Trigger Word: remind
Request URL: https://<backend-public-host>/mattermost/commands/remind
Method: POST
Usage: /remind <No> <minutes>
```

例:

```text
/remind 7 30
```

`No` は Notion の display ID。`TASK-7` は `7` でも一致する。
`minutes` は現在時刻から何分後に再通知するか。

### create

```text
Trigger Word: create
Request URL: https://<backend-public-host>/mattermost/commands/create
Method: POST
Usage: /create <title> <date_mmddhhmm> <notification_mmddhhmm>
```

例:

```text
/create write report 08100930 08101000
```

末尾2つの token を `date` と `notification` として扱い、それより前を title にする。
日時 token は `mmddhhmm` 形式で、現在年として扱う。
日時 token に `-` を入れると、その値は未指定として Notion に送らない。

### quick

```text
Trigger Word: quick
Request URL: https://<backend-public-host>/mattermost/commands/quick
Method: POST
Usage: /quick <title> <date_mmddhhmm> <notification_mmddhhmm>
```

例:

```text
/quick call customer 08100930 08101000
```

`create` と同じ形式で Notion item を作る。
違いは priority を `High` にすること。
`date` / `notification` は `-` で未指定にできる。

### work

```text
Trigger Word: work
Request URL: https://<backend-public-host>/mattermost/commands/work
Method: POST
Usage: /work <start|end|todo> <start_mmddhhmm> <end_mmddhhmm>
```

例:

```text
/work start 08100930 08111845
```

`mmddhhmm` は現在年の日時として扱う。
`end_mmddhhmm` は `-` で未指定にできる。`start_mmddhhmm` と `end_mmddhhmm` の両方を `-` にすると date を未指定にする。

```text
08100930 -> YYYY-08-10 09:30
08111845 -> YYYY-08-11 18:45
```

status は以下に変換する。

```text
start -> inprogress
end   -> done
todo  -> todo
```

`/work` は `NOTION_WORK_TEMPLATE_ID` の template を使って Notion item を作る。

```dotenv
NOTION_WORK_TEMPLATE_ID=
```

## Notes

Mattermost 側では slash command を4つ登録する。
backend public host が `med-control.med-000.dev` の場合、Request URL は以下になる。

```text
https://med-control.med-000.dev/mattermost/commands/remind
https://med-control.med-000.dev/mattermost/commands/create
https://med-control.med-000.dev/mattermost/commands/quick
https://med-control.med-000.dev/mattermost/commands/work
```
