# データベース設計書

## ER図

```
┌─────────────┐       ┌──────────────────┐       ┌─────────────────┐
│   staffs    │       │ shift_requests   │       │  constraints    │
│─────────────│       │──────────────────│       │─────────────────│
│ id (PK)     │──┐    │ id (PK)          │       │ id (PK)         │
│ name        │  ├───<│ staff_id (FK)    │       │ name            │
│ role        │  │    │ year_month       │       │ type            │
│ employment_ │  │    │ date             │       │ category        │
│   type      │  │    │ start_time       │       │ config (JSONB)  │
│ is_active   │  │    │ end_time         │       │ is_active       │
│ created_at  │  │    │ request_type     │       │ priority        │
│ updated_at  │  │    │ note             │       │ created_at      │
└─────────────┘  │    │ created_at       │       │ updated_at      │
                 │    │ updated_at       │       └─────────────────┘
                 │    └──────────────────┘
                 │
                 │    ┌──────────────────────┐
                 │    │ staff_monthly_       │
                 │    │   settings           │
                 ├───<│──────────────────────│
                 │    │ id (PK)              │
                 │    │ staff_id (FK)        │
                 │    │ year_month           │
                 │    │ min_preferred_hours  │
                 │    │ max_preferred_hours  │
                 │    │ note                 │
                 │    │ created_at           │
                 │    │ updated_at           │
                 │    └──────────────────────┘
                 │
                 │    ┌──────────────────┐       ┌─────────────────┐
                 │    │ shift_patterns   │       │  shift_entries  │
                 │    │──────────────────│       │─────────────────│
                 │    │ id (PK)          │──────<│ id (PK)         │
                 │    │ year_month       │       │ pattern_id (FK) │
                 │    │ status           │       │ staff_id (FK)───┤
                 │    │ reasoning        │       │ date            │
                 │    │ score            │       │ start_time      │
                 │    │ constraint_      │       │ end_time        │
                 │    │  violations      │       │ break_minutes   │
                 │    │ created_at       │       │ is_manual_edit  │
                 │    │ updated_at       │       │ created_at      │
                 │    └──────────────────┘       │ updated_at      │
                 │                               └─────────────────┘
                 └───────────────────────────────────────┘
```

## テーブル定義

### staffs（スタッフ）

| カラム | 型 | NOT NULL | デフォルト | 説明 |
|--------|-----|----------|-----------|------|
| id | UUID | YES | gen_random_uuid() | 主キー |
| name | VARCHAR(100) | YES | - | 氏名 |
| role | VARCHAR(50) | YES | - | 役割（kitchen/hall/cleaning等） |
| employment_type | VARCHAR(20) | YES | - | 雇用形態（full_time/part_time） |
| is_active | BOOLEAN | YES | true | 有効フラグ |
| created_at | TIMESTAMPTZ | YES | NOW() | 作成日時 |
| updated_at | TIMESTAMPTZ | YES | NOW() | 更新日時 |

### staff_monthly_settings（スタッフ月間設定）

毎月のスタッフごとの希望労働時間を管理する。
従業員から受け取った「今月は○○〜○○時間働きたい」という希望を記録する。

| カラム | 型 | NOT NULL | デフォルト | 説明 |
|--------|-----|----------|-----------|------|
| id | UUID | YES | gen_random_uuid() | 主キー |
| staff_id | UUID | YES | - | FK: staffs.id |
| year_month | VARCHAR(7) | YES | - | 対象年月（"2026-03"形式） |
| min_preferred_hours | INTEGER | YES | - | 最低希望労働時間（時間） |
| max_preferred_hours | INTEGER | YES | - | 最大希望労働時間（時間） |
| note | TEXT | NO | NULL | 備考（扶養控除の制限等） |
| created_at | TIMESTAMPTZ | YES | NOW() | 作成日時 |
| updated_at | TIMESTAMPTZ | YES | NOW() | 更新日時 |

**ユニーク制約:** `(staff_id, year_month)` — 同一スタッフ・同一月で1レコード

### shift_requests（シフト希望）

| カラム | 型 | NOT NULL | デフォルト | 説明 |
|--------|-----|----------|-----------|------|
| id | UUID | YES | gen_random_uuid() | 主キー |
| staff_id | UUID | YES | - | FK: staffs.id |
| year_month | VARCHAR(7) | YES | - | 対象年月（"2026-03"形式） |
| date | DATE | YES | - | 希望日 |
| start_time | TIME | NO | NULL | 希望開始時刻（NULLなら終日） |
| end_time | TIME | NO | NULL | 希望終了時刻 |
| request_type | VARCHAR(20) | YES | - | available/unavailable/preferred |
| note | TEXT | NO | NULL | 備考 |
| created_at | TIMESTAMPTZ | YES | NOW() | 作成日時 |
| updated_at | TIMESTAMPTZ | YES | NOW() | 更新日時 |

**request_type の意味:**
- `available`: 出勤可能
- `unavailable`: 出勤不可（ハード制約として扱う）
- `preferred`: できれば出勤したい（ソフト制約として扱う）

### constraints（制約条件）

| カラム | 型 | NOT NULL | デフォルト | 説明 |
|--------|-----|----------|-----------|------|
| id | UUID | YES | gen_random_uuid() | 主キー |
| name | VARCHAR(200) | YES | - | 制約名 |
| type | VARCHAR(10) | YES | - | hard/soft |
| category | VARCHAR(50) | YES | - | 制約カテゴリ |
| config | JSONB | YES | '{}' | 制約パラメータ |
| is_active | BOOLEAN | YES | true | 有効フラグ |
| priority | INTEGER | NO | 0 | 優先度（ソフト制約用、高いほど重要） |
| created_at | TIMESTAMPTZ | YES | NOW() | 作成日時 |
| updated_at | TIMESTAMPTZ | YES | NOW() | 更新日時 |

**category と config の例:**

```json
// category: "min_staff" - 最低スタッフ数
{
  "min_count": 2,
  "time_ranges": [
    {"start": "10:00", "end": "14:00", "min_count": 3},
    {"start": "17:00", "end": "21:00", "min_count": 3}
  ]
}

// category: "max_consecutive_days" - 連勤制限
{
  "max_days": 5
}

// category: "monthly_hours" - 月間労働時間制限
{
  "min_hours": 60,
  "max_hours": 160
}

// category: "fixed_day_off" - 固定休日
{
  "staff_id": "uuid-here",
  "day_of_week": 3  // 水曜
}

// category: "staff_compatibility" - スタッフ相性
{
  "staff_ids": ["uuid-1", "uuid-2"],
  "rule": "prefer_together"  // prefer_together / avoid_together
}

// category: "rest_hours" - 勤務間インターバル
{
  "min_hours": 11
}

// category: "max_staff" - 最大スタッフ数（人件費制約）
{
  "max_count": 5
}
```

### shift_patterns（シフトパターン）

| カラム | 型 | NOT NULL | デフォルト | 説明 |
|--------|-----|----------|-----------|------|
| id | UUID | YES | gen_random_uuid() | 主キー |
| year_month | VARCHAR(7) | YES | - | 対象年月 |
| status | VARCHAR(20) | YES | 'draft' | draft/selected/finalized |
| reasoning | TEXT | NO | NULL | LLMの生成理由説明 |
| score | DECIMAL(5,2) | NO | NULL | パターン品質スコア（0-100） |
| constraint_violations | JSONB | NO | '[]' | ソフト制約違反の一覧 |
| created_at | TIMESTAMPTZ | YES | NOW() | 作成日時 |
| updated_at | TIMESTAMPTZ | YES | NOW() | 更新日時 |

**status の遷移:**
```
draft → selected → finalized
         ↓
       draft（選択解除時）
```

### shift_entries（シフトエントリ）

| カラム | 型 | NOT NULL | デフォルト | 説明 |
|--------|-----|----------|-----------|------|
| id | UUID | YES | gen_random_uuid() | 主キー |
| pattern_id | UUID | YES | - | FK: shift_patterns.id |
| staff_id | UUID | YES | - | FK: staffs.id |
| date | DATE | YES | - | シフト日 |
| start_time | TIME | YES | - | 開始時刻 |
| end_time | TIME | YES | - | 終了時刻 |
| break_minutes | INTEGER | YES | 0 | 休憩時間（分） |
| is_manual_edit | BOOLEAN | YES | false | 手動編集フラグ |
| created_at | TIMESTAMPTZ | YES | NOW() | 作成日時 |
| updated_at | TIMESTAMPTZ | YES | NOW() | 更新日時 |

### generation_jobs（生成ジョブ）

| カラム | 型 | NOT NULL | デフォルト | 説明 |
|--------|-----|----------|-----------|------|
| id | UUID | YES | gen_random_uuid() | 主キー |
| year_month | VARCHAR(7) | YES | - | 対象年月 |
| status | VARCHAR(20) | YES | 'pending' | pending/processing/completed/failed |
| pattern_count | INTEGER | YES | 3 | 生成パターン数 |
| error_message | TEXT | NO | NULL | エラーメッセージ |
| started_at | TIMESTAMPTZ | NO | NULL | 処理開始日時 |
| completed_at | TIMESTAMPTZ | NO | NULL | 処理完了日時 |
| created_at | TIMESTAMPTZ | YES | NOW() | 作成日時 |

## インデックス

```sql
-- staff_monthly_settings
CREATE UNIQUE INDEX idx_staff_monthly_settings_unique ON staff_monthly_settings(staff_id, year_month);

-- shift_requests
CREATE INDEX idx_shift_requests_staff_month ON shift_requests(staff_id, year_month);
CREATE INDEX idx_shift_requests_year_month ON shift_requests(year_month);

-- shift_patterns
CREATE INDEX idx_shift_patterns_year_month ON shift_patterns(year_month);
CREATE INDEX idx_shift_patterns_status ON shift_patterns(status);

-- shift_entries
CREATE INDEX idx_shift_entries_pattern ON shift_entries(pattern_id);
CREATE INDEX idx_shift_entries_staff_date ON shift_entries(staff_id, date);

-- constraints
CREATE INDEX idx_constraints_active ON constraints(is_active);
CREATE INDEX idx_constraints_category ON constraints(category);

-- generation_jobs
CREATE INDEX idx_generation_jobs_status ON generation_jobs(status);
CREATE INDEX idx_generation_jobs_year_month ON generation_jobs(year_month);
```

## マイグレーション

マイグレーションツールは `golang-migrate/migrate` を使用。
マイグレーションファイルは `backend/migrations/` に配置。
