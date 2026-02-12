# API設計書

## 共通仕様

- **ベースURL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`
- **エラーレスポンス形式**:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "名前は必須です",
    "details": [
      {"field": "name", "message": "名前は必須です"}
    ]
  }
}
```

## エンドポイント一覧

### スタッフ管理

#### `GET /api/v1/staffs`
スタッフ一覧取得

**クエリパラメータ:**
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| is_active | boolean | NO | 有効フラグでフィルタ |

**レスポンス: 200**
```json
{
  "staffs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "田中太郎",
      "role": "kitchen",
      "employment_type": "full_time",
      "is_active": true,
      "created_at": "2026-01-15T09:00:00Z",
      "updated_at": "2026-01-15T09:00:00Z"
    }
  ]
}
```

#### `POST /api/v1/staffs`
スタッフ登録

**リクエスト:**
```json
{
  "name": "田中太郎",
  "role": "kitchen",
  "employment_type": "full_time"
}
```

**レスポンス: 201** — 作成されたスタッフオブジェクト

#### `GET /api/v1/staffs/:id`
スタッフ詳細取得

**レスポンス: 200** — スタッフオブジェクト

#### `PUT /api/v1/staffs/:id`
スタッフ更新

**リクエスト:** POST と同じ形式（部分更新可）

**レスポンス: 200** — 更新後のスタッフオブジェクト

#### `DELETE /api/v1/staffs/:id`
スタッフ削除（論理削除: is_active = false）

**レスポンス: 204** No Content

---

### スタッフ月間設定

#### `GET /api/v1/staff-monthly-settings`
月間設定一覧取得

**クエリパラメータ:**
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| year_month | string | YES | 対象年月（"2026-03"） |
| staff_id | UUID | NO | スタッフでフィルタ |

**レスポンス: 200**
```json
{
  "settings": [
    {
      "id": "...",
      "staff_id": "...",
      "staff_name": "田中太郎",
      "year_month": "2026-03",
      "min_preferred_hours": 80,
      "max_preferred_hours": 120,
      "note": "扶養控除のため月120h以内",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

#### `POST /api/v1/staff-monthly-settings`
月間設定登録（同一スタッフ・同一月で既存がある場合は上書き）

**リクエスト:**
```json
{
  "staff_id": "550e8400-e29b-41d4-a716-446655440000",
  "year_month": "2026-03",
  "min_preferred_hours": 80,
  "max_preferred_hours": 120,
  "note": ""
}
```

**レスポンス: 201**

#### `POST /api/v1/staff-monthly-settings/batch`
月間設定一括登録（全スタッフ分をまとめて登録）

**リクエスト:**
```json
{
  "settings": [
    {
      "staff_id": "...",
      "year_month": "2026-03",
      "min_preferred_hours": 80,
      "max_preferred_hours": 120
    }
  ]
}
```

**レスポンス: 201**
```json
{
  "created_count": 5,
  "settings": [...]
}
```

#### `PUT /api/v1/staff-monthly-settings/:id`
月間設定更新

**レスポンス: 200**

#### `DELETE /api/v1/staff-monthly-settings/:id`
月間設定削除

**レスポンス: 204**

---

### シフト希望

#### `GET /api/v1/shift-requests`
シフト希望一覧取得

**クエリパラメータ:**
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| year_month | string | YES | 対象年月（"2026-03"） |
| staff_id | UUID | NO | スタッフでフィルタ |

**レスポンス: 200**
```json
{
  "shift_requests": [
    {
      "id": "...",
      "staff_id": "...",
      "staff_name": "田中太郎",
      "year_month": "2026-03",
      "date": "2026-03-15",
      "start_time": "09:00",
      "end_time": "17:00",
      "request_type": "available",
      "note": "午前中希望",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

#### `POST /api/v1/shift-requests`
シフト希望登録（単件）

**リクエスト:**
```json
{
  "staff_id": "550e8400-e29b-41d4-a716-446655440000",
  "year_month": "2026-03",
  "date": "2026-03-15",
  "start_time": "09:00",
  "end_time": "17:00",
  "request_type": "available",
  "note": ""
}
```

**レスポンス: 201**

#### `POST /api/v1/shift-requests/batch`
シフト希望一括登録

**リクエスト:**
```json
{
  "requests": [
    {
      "staff_id": "...",
      "year_month": "2026-03",
      "date": "2026-03-15",
      "start_time": "09:00",
      "end_time": "17:00",
      "request_type": "available"
    }
  ]
}
```

**レスポンス: 201**
```json
{
  "created_count": 5,
  "shift_requests": [...]
}
```

#### `PUT /api/v1/shift-requests/:id`
シフト希望更新

**レスポンス: 200**

#### `DELETE /api/v1/shift-requests/:id`
シフト希望削除

**レスポンス: 204**

---

### 制約条件

#### `GET /api/v1/constraints`
制約一覧取得

**クエリパラメータ:**
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| is_active | boolean | NO | 有効な制約のみ |
| type | string | NO | hard/soft でフィルタ |
| category | string | NO | カテゴリでフィルタ |

**レスポンス: 200**
```json
{
  "constraints": [
    {
      "id": "...",
      "name": "ランチタイム最低3名",
      "type": "hard",
      "category": "min_staff",
      "config": {
        "time_ranges": [
          {"start": "11:00", "end": "14:00", "min_count": 3}
        ]
      },
      "is_active": true,
      "priority": 0,
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

#### `POST /api/v1/constraints`
制約作成

**リクエスト:**
```json
{
  "name": "連勤制限5日",
  "type": "hard",
  "category": "max_consecutive_days",
  "config": {
    "max_days": 5
  },
  "priority": 0
}
```

**レスポンス: 201**

#### `PUT /api/v1/constraints/:id`
制約更新

**レスポンス: 200**

#### `DELETE /api/v1/constraints/:id`
制約削除

**レスポンス: 204**

---

### ダッシュボード

#### `GET /api/v1/dashboard/summary`
ダッシュボード用サマリー取得

**クエリパラメータ:**
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| year_month | string | YES | 対象年月（"2026-03"） |

**レスポンス: 200**
```json
{
  "year_month": "2026-03",
  "staff_count": 8,
  "active_staff_count": 8,
  "request_submitted_count": 6,
  "monthly_settings_count": 7,
  "shift_status": "generated",
  "constraint_count": 5,
  "daily_staff_counts": [
    {"date": "2026-03-01", "count": 5},
    {"date": "2026-03-02", "count": 4}
  ]
}
```

**shift_status の値:**
- `not_started`: シフト希望未入力
- `requests_submitted`: 希望入力済み、未生成
- `generating`: 生成中
- `generated`: 生成完了（パターン選択前）
- `selected`: パターン選択済み（未確定）
- `finalized`: 確定済み

---

### シフト生成

#### `POST /api/v1/shifts/generate`
シフト自動生成（非同期ジョブ開始）

**リクエスト:**
```json
{
  "year_month": "2026-03",
  "pattern_count": 3
}
```

**レスポンス: 202**
```json
{
  "job_id": "...",
  "status": "pending",
  "message": "シフト生成を開始しました"
}
```

#### `GET /api/v1/shifts/generate/:job_id`
生成ジョブ状態確認

**レスポンス: 200**
```json
{
  "job_id": "...",
  "status": "processing",
  "year_month": "2026-03",
  "pattern_count": 3,
  "started_at": "2026-03-01T10:00:00Z",
  "completed_at": null,
  "error_message": null
}
```

**status の遷移:** `pending → processing → completed / failed`

---

### シフトパターン

#### `GET /api/v1/shifts/patterns`
生成されたパターン一覧

**クエリパラメータ:**
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| year_month | string | YES | 対象年月 |

**レスポンス: 200**
```json
{
  "patterns": [
    {
      "id": "...",
      "year_month": "2026-03",
      "status": "draft",
      "reasoning": "パターン1は全員の希望をできる限り反映し...",
      "score": 85.5,
      "constraint_violations": [
        {
          "constraint_name": "Aさん週3希望",
          "type": "soft",
          "message": "Aさんの週3希望に対し、第2週は2日のみ"
        }
      ],
      "summary": {
        "_note": "shift_entries から算出される計算値（DB には保存しない）",
        "total_entries": 45,
        "staff_hours": {
          "田中太郎": 120,
          "佐藤花子": 80
        }
      },
      "created_at": "..."
    }
  ]
}
```

#### `GET /api/v1/shifts/patterns/:id`
パターン詳細（エントリ含む）

**レスポンス: 200**
```json
{
  "pattern": {
    "id": "...",
    "year_month": "2026-03",
    "status": "draft",
    "reasoning": "...",
    "score": 85.5,
    "constraint_violations": [...],
    "entries": [
      {
        "id": "...",
        "staff_id": "...",
        "staff_name": "田中太郎",
        "date": "2026-03-01",
        "start_time": "09:00",
        "end_time": "17:00",
        "break_minutes": 60,
        "is_manual_edit": false
      }
    ]
  }
}
```

#### `PUT /api/v1/shifts/patterns/:id/select`
パターン選択

**レスポンス: 200**
```json
{
  "pattern": {
    "id": "...",
    "status": "selected"
  }
}
```

#### `PUT /api/v1/shifts/patterns/:id/finalize`
パターン確定

**レスポンス: 200**
```json
{
  "pattern": {
    "id": "...",
    "status": "finalized"
  }
}
```

---

### シフトエントリ編集

#### `PUT /api/v1/shifts/entries/:id`
個別シフトエントリ編集

**リクエスト:**
```json
{
  "start_time": "10:00",
  "end_time": "18:00",
  "break_minutes": 60
}
```

**レスポンス: 200** — 更新後のエントリ + バリデーション結果

```json
{
  "entry": {...},
  "validation": {
    "is_valid": true,
    "warnings": [
      {
        "type": "soft_constraint",
        "message": "この変更により田中さんの月間労働時間が上限に近づきます（155/160h）"
      }
    ]
  }
}
```

#### `POST /api/v1/shifts/entries`
シフトエントリ追加（パターン内に新規エントリ追加）

**リクエスト:**
```json
{
  "pattern_id": "...",
  "staff_id": "...",
  "date": "2026-03-20",
  "start_time": "09:00",
  "end_time": "17:00",
  "break_minutes": 60
}
```

**レスポンス: 201**

#### `DELETE /api/v1/shifts/entries/:id`
シフトエントリ削除

**レスポンス: 204**

---

### PDF出力

PDF生成はフロントエンドで実行（jsPDF）。バックエンドからはパターン詳細 API で必要なデータを取得する。

**フロントエンドでの生成フロー:**
1. `GET /api/v1/shifts/patterns/:id` でデータ取得
2. jsPDF + jspdf-autotable でテーブル形式の PDF を生成
3. ブラウザでダウンロード
