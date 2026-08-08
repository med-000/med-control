# Database

`med-control` の SQLite は backend が所有する。
Notion は人間が管理する正本、SQLite は med-control が動くための local snapshot と制御状態を持つ。

## Tables

```sql
items (
  notion_page_id TEXT PRIMARY KEY,
  item_id TEXT NOT NULL UNIQUE,
  display_id TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  status_id TEXT,
  status_name TEXT,
  status_color TEXT,
  date_start TEXT,
  date_end TEXT,
  date_time_zone TEXT,
  label_id TEXT,
  label_name TEXT,
  label_color TEXT,
  priority_id TEXT,
  priority_name TEXT,
  priority_color TEXT,
  notification_start TEXT,
  notification_end TEXT,
  notification_time_zone TEXT,
  source_url TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  synced_at TEXT NOT NULL,
  last_seen_at TEXT NOT NULL,
  deleted_at TEXT
);
```

```sql
item_categories (
  notion_page_id TEXT NOT NULL,
  category_id TEXT NOT NULL DEFAULT '',
  category_name TEXT NOT NULL,
  category_color TEXT NOT NULL DEFAULT '',
  position INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (notion_page_id, category_id, category_name),
  FOREIGN KEY (notion_page_id) REFERENCES items(notion_page_id) ON DELETE CASCADE
);
```

```sql
item_control_state (
  notion_page_id TEXT PRIMARY KEY,
  reminder_override_start TEXT,
  reminder_override_end TEXT,
  reminder_override_time_zone TEXT,
  reminder_updated_at TEXT,
  reminder_source TEXT NOT NULL DEFAULT '',
  FOREIGN KEY (notion_page_id) REFERENCES items(notion_page_id) ON DELETE CASCADE
);
```

```sql
sent_notifications (
  notification_key TEXT PRIMARY KEY,
  notion_page_id TEXT,
  task_id TEXT NOT NULL,
  display_id TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  notification_at TEXT,
  sent_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (notion_page_id) REFERENCES items(notion_page_id) ON DELETE SET NULL
);
```

## Ownership

- `items`: Notion page の local snapshot。task / idea / memo などを広く item と呼ぶ。
- `item_categories`: Notion multi-select category の正規化 table。
- `item_control_state`: `/remind` など、Notion DB に持たせない med-control 側の状態。
- `sent_notifications`: Mattermost への二重通知防止履歴。

`notion_page_id` を relation の中心にする。
domain の `Task.ID` 互換用に `item_id` には `notion:<page_id>` を保存する。
