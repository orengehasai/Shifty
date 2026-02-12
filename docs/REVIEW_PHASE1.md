# Phase 1 設計レビューレポート

## サマリー
- チェック項目数: 12
- 問題あり: 3件
- 警告: 4件
- 問題なし: 5件

---

## 問題一覧

### [重要度: HIGH] スタッフ管理画面の「月間上限H」が DB・API と不整合

- **対象ファイル**: docs/SCREENS.md (スタッフ管理画面), docs/DATABASE.md (staffs テーブル), docs/API.md (スタッフ API)
- **内容**: スタッフ管理画面のテーブルに「月間上限H」列があり、スタッフ登録/編集モーダルにも「月間上限: [160] 時間」の入力欄がある。しかし、`staffs` テーブルにはこのカラムが存在せず、月間労働時間の情報は `staff_monthly_settings` テーブルで月ごとに管理する設計になっている。スタッフの属性として固定の上限時間を持つのか、月ごとに変動する設定なのかが設計書間で矛盾している。
- **修正案**:
  - **案A（推奨）**: スタッフ管理画面から「月間上限H」列と入力欄を削除する。月間希望労働時間はシフト希望入力画面（`/requests`）で `staff_monthly_settings` として管理する（現在の DB 設計に合わせる）。
  - **案B**: `staffs` テーブルに `default_max_monthly_hours INTEGER` カラムを追加し、スタッフ登録時にデフォルト値を設定できるようにする。`staff_monthly_settings` が未設定の月はこのデフォルト値を使用する。

  ```markdown
  # SCREENS.md のスタッフ管理テーブルを以下に修正（案A の場合）:
  │  ┌─────┬────────┬─────┬──────┬─────────┐ │
  │  │ 名前 │ 役割   │ 雇用 │ 状態 │ 操作    │ │
  │  ├─────┼────────┼─────┼──────┼─────────┤ │
  │  │田中  │キッチン │正社員│ 有効 │ 編集|削除│ │
  │  │佐藤  │ホール  │パート│ 有効 │ 編集|削除│ │
  │  └─────┴────────┴─────┴──────┴─────────┘ │

  # スタッフ登録/編集モーダルから「月間上限」入力欄を削除
  ```

---

### [重要度: HIGH] LLM 出力の constraint_violations フォーマットが API レスポンスと不一致

- **対象ファイル**: docs/LLM_INTEGRATION.md (出力JSON形式), docs/API.md (shift_patterns レスポンス)
- **内容**: LLM の出力 JSON では `soft_constraint_notes` として `{ "constraint_name": "...", "note": "..." }` の形式で出力するが、API のパターン一覧レスポンスの `constraint_violations` は `{ "constraint_name": "...", "type": "soft", "message": "..." }` という形式になっている。フィールド名が異なる（`note` vs `message`）のに加え、API 側には `type` フィールドがあるが LLM 出力にはない。バックエンドで変換処理を入れるか、どちらかの形式に統一する必要がある。
- **修正案**: LLM 出力フォーマットを API レスポンスの形式に合わせる。バックエンドでの変換処理を不要にし、LLM 出力を直接 `shift_patterns.constraint_violations` (JSONB) に保存可能にする。

  ```markdown
  # LLM_INTEGRATION.md のシステムプロンプト内 JSON 形式を以下に修正:

  ## 出力JSON形式
  {
    "reasoning": "このパターンの特徴と判断理由の説明",
    "entries": [
      {
        "staff_id": "uuid",
        "date": "2026-03-01",
        "start_time": "09:00",
        "end_time": "17:00",
        "break_minutes": 60
      }
    ],
    "constraint_violations": [
      {
        "constraint_name": "制約名",
        "type": "soft",
        "message": "この制約を完全には満たせなかった理由"
      }
    ]
  }
  ```

---

### [重要度: HIGH] shift_patterns API レスポンスの summary フィールドに対応する DB 設計がない

- **対象ファイル**: docs/API.md (GET /api/v1/shifts/patterns), docs/DATABASE.md (shift_patterns テーブル)
- **内容**: `GET /api/v1/shifts/patterns` のレスポンスに `summary` オブジェクト（`total_entries`, `staff_hours`）が含まれているが、`shift_patterns` テーブルにはこのデータを格納するカラムがない。これがサーバー側で毎回算出する計算値なのか、DB にキャッシュするのかが明確でない。
- **修正案**: `summary` はパフォーマンスへの影響が小さいため（エントリ数の集計とスタッフ別時間計算）、サーバー側で `shift_entries` テーブルから都度算出するのが適切。DB にカラムを追加する必要はないが、API 設計書にこれが計算値であることを明記すべき。

  ```markdown
  # API.md のパターン一覧レスポンスに注記を追加:

  "summary": {
    // ※ shift_entries から算出される計算値（DB には保存しない）
    "total_entries": 45,
    "staff_hours": {
      "田中太郎": 120,
      "佐藤花子": 80
    }
  }
  ```

---

### [重要度: MEDIUM] ダッシュボード画面用のサマリー API が未定義

- **対象ファイル**: docs/SCREENS.md (ダッシュボード画面), docs/API.md
- **内容**: ダッシュボード画面では「スタッフ数」「希望提出率（6/8）」「シフト状態（生成済み等）」「制約数」および「日付ごとの出勤予定人数」を表示する。これらのデータを取得するには `GET /staffs`, `GET /shift-requests`, `GET /shifts/patterns`, `GET /constraints` を個別に呼び出し、フロント側で集計する必要がある。ダッシュボード表示のたびに4つ以上の API を呼ぶことになり非効率。
- **修正案**: ダッシュボード用のサマリー API エンドポイントを追加する。

  ```markdown
  # API.md に以下のエンドポイントを追加:

  ### ダッシュボード

  #### `GET /api/v1/dashboard/summary`
  ダッシュボード用サマリー取得

  **クエリパラメータ:**
  | パラメータ | 型 | 必須 | 説明 |
  |-----------|-----|------|------|
  | year_month | string | YES | 対象年月 |

  **レスポンス: 200**
  {
    "year_month": "2026-03",
    "staff_count": 8,
    "active_staff_count": 8,
    "request_submitted_count": 6,
    "shift_status": "generated",
    "constraint_count": 5,
    "daily_staff_counts": [
      {"date": "2026-03-01", "count": 5},
      {"date": "2026-03-02", "count": 4}
    ]
  }
  ```

---

### [重要度: MEDIUM] シフト希望入力画面の「月間希望時間」が単一値として表示されている

- **対象ファイル**: docs/SCREENS.md (シフト希望入力画面), docs/DATABASE.md (staff_monthly_settings テーブル)
- **内容**: シフト希望入力画面では「月間希望時間: [120] 時間」と単一値の入力欄になっているが、DB 設計では `min_preferred_hours` と `max_preferred_hours` のレンジ（範囲）で管理している。画面設計と DB 設計で希望時間の表現方法が異なっている。
- **修正案**: 画面設計を DB 設計に合わせて、最小〜最大のレンジ入力に変更する。

  ```markdown
  # SCREENS.md のシフト希望入力画面を修正:
  │  月間希望時間: [80] 〜 [120] 時間                       │
  ```

---

### [重要度: MEDIUM] ARCHITECTURE.md のシフト生成フロー中の API パスが API.md と不一致

- **対象ファイル**: docs/ARCHITECTURE.md (シフト生成フロー), docs/API.md
- **内容**: ARCHITECTURE.md の通信フロー図では `GET /status` と `GET /patterns` というパスが記載されているが、API.md では `GET /api/v1/shifts/generate/:job_id`（ジョブ状態確認）と `GET /api/v1/shifts/patterns`（パターン一覧）が正式なパスとなっている。ARCHITECTURE.md の記載が省略形で記載されているため、誤解を招く。
- **修正案**: ARCHITECTURE.md のフロー図に正式な API パスを記載する。

  ```markdown
  # ARCHITECTURE.md のシフト生成フローを修正:
  │── GET /api/v1/shifts/generate/:job_id ─►│
  │◄── {status: processing} ────────────────│
  ...
  │── GET /api/v1/shifts/generate/:job_id ─►│
  │◄── {status: completed} ─────────────────│
  │── GET /api/v1/shifts/patterns?year_month=... ─►│
  │◄── [pattern1, pattern2..] ──────────────│
  ```

---

### [重要度: LOW] generation_jobs テーブルの pattern_count と created_at が API レスポンスに含まれていない

- **対象ファイル**: docs/API.md (GET /api/v1/shifts/generate/:job_id), docs/DATABASE.md (generation_jobs テーブル)
- **内容**: `generation_jobs` テーブルには `pattern_count` と `created_at` カラムがあるが、ジョブ状態確認の API レスポンスにはこれらのフィールドが含まれていない。フロントエンドで生成パターン数を表示する場合に情報が不足する可能性がある。
- **修正案**: API レスポンスに `pattern_count` を追加する。

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

---

## 問題なしの項目

| # | チェック内容 | 結果 |
|---|------------|------|
| 1 | staffs テーブル <-> スタッフ CRUD API のフィールド整合 | OK |
| 2 | staff_monthly_settings テーブル <-> 月間設定 API のフィールド整合 | OK |
| 3 | shift_requests テーブル <-> シフト希望 API のフィールド整合 | OK |
| 4 | constraints テーブル <-> 制約 API のフィールド整合 | OK |
| 5 | 命名規則の統一（API パスはケバブケース、DB はスネークケース、コンポーネントはパスカルケース） | OK |

---

## 推奨対応優先度

| 優先度 | 問題 | 影響 |
|--------|------|------|
| 1 | スタッフ管理画面の「月間上限H」の不整合 | 実装時にフロントとバックで認識が齟齬する |
| 2 | LLM 出力の constraint_violations フォーマット不一致 | LLM連携実装時に変換処理の追加が必要になる |
| 3 | shift_patterns summary フィールドの DB 設計不足 | 実装方針が曖昧になる |
| 4 | ダッシュボード用サマリー API の追加 | フロント実装時に API 不足で手戻りが発生する |
| 5 | 月間希望時間の単一値 vs レンジの不一致 | UI 実装時に混乱する |
| 6 | ARCHITECTURE.md のフロー図の API パス修正 | 軽微だが正確性のため修正推奨 |
| 7 | generation_jobs の pattern_count 追加 | 軽微。フロントで必要になった際に追加すればよい |
