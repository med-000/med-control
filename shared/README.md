# shared

backend と infra が共有する domain 型を置く module。

## 構成

```text
domain/task
```

## 方針

- backend と infra の受け渡しに必要な型だけ置く
- Notion API の生レスポンス型は置かない
- backend の DB model や usecase は置かない
- 外部サービス固有の実装は置かない

`Task`, `SelectOption`, `DateRange` など、backend が直感的に扱うための型をここに置く。
