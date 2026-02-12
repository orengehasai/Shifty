# Phase 3 設計書 vs 実装 照合レポート

## サマリー
- チェック項目数: 28
- 一致: 22件
- 差分あり（修正済み）: 5件
- 差分あり（設計変更推奨）: 1件

---

## 1. API エンドポイント照合（docs/API.md vs 実装）

### [一致] GET /api/v1/staffs
- ルーティング: `staff_handler.go:21` で定義
- クエリパラメータ `is_active` 対応済み
- レスポンス形式 `{ staffs: [...] }` 一致

### [一致] POST /api/v1/staffs
- ルーティング: `staff_handler.go:22`
- リクエスト形式（name, role, employment_type）一致
- レスポンス 201 一致

### [一致] GET /api/v1/staffs/:id
- ルーティング: `staff_handler.go:23` で定義

### [一致] PUT /api/v1/staffs/:id
- ルーティング: `staff_handler.go:24` で定義

### [一致] DELETE /api/v1/staffs/:id
- ルーティング: `staff_handler.go:25` で定義、204 NoContent 返却

### [一致] GET /api/v1/staff-monthly-settings
- ルーティング: `staff_monthly_setting_handler.go:21`
- year_month 必須チェック、staff_id フィルタ対応済み

### [一致] POST /api/v1/staff-monthly-settings
- ルーティング: `staff_monthly_setting_handler.go:22`

### [一致] POST /api/v1/staff-monthly-settings/batch
- ルーティング: `staff_monthly_setting_handler.go:23`
- レスポンス `{ created_count, settings }` 一致

### [一致] PUT /api/v1/staff-monthly-settings/:id
- ルーティング: `staff_monthly_setting_handler.go:24`

### [一致] DELETE /api/v1/staff-monthly-settings/:id
- ルーティング: `staff_monthly_setting_handler.go:25`

### [一致] GET /api/v1/shift-requests, POST, batch, PUT, DELETE
- 全エンドポイント実装済み、レスポンス形式一致

### [一致] GET /api/v1/constraints, POST, PUT, DELETE
- 全エンドポイント実装済み、フィルタパラメータ(is_active, type, category)対応

### [一致] GET /api/v1/dashboard/summary
- ルーティング: `dashboard_handler.go:20`
- year_month 必須チェック対応
- レスポンス形式（DashboardSummary）一致

### [一致] POST /api/v1/shifts/generate
- 202 Accepted レスポンス、非同期ジョブ開始

### [一致] GET /api/v1/shifts/generate/:job_id
- ジョブステータス確認 API 実装済み

### [一致] GET /api/v1/shifts/patterns
- パターン一覧 + summary 計算値対応

### [一致] GET /api/v1/shifts/patterns/:id
- パターン詳細（entries 含む）レスポンス一致

### [一致] PUT /api/v1/shifts/patterns/:id/select, /finalize
- パターン選択・確定 API 実装済み

### [一致] POST /api/v1/shifts/entries, PUT, DELETE
- シフトエントリ CRUD 全実装済み
- PUT のレスポンスに `{ entry, validation }` 形式対応

---

## 2. DB スキーマ照合（docs/DATABASE.md vs 実装）

### [一致] staffs テーブル
- マイグレーション `000001_create_tables.up.sql` と DATABASE.md のカラム定義が完全一致
- モデル `model.Staff` のフィールドも一致

### [一致] staff_monthly_settings テーブル
- ユニーク制約 `(staff_id, year_month)` 実装済み
- min_preferred_hours, max_preferred_hours カラム一致

### [一致] shift_requests テーブル
- start_time, end_time が NULL 許容（NOT NULL でない）
- request_type: available/unavailable/preferred 対応

### [一致] constraints テーブル
- config JSONB, priority INTEGER 一致

### [一致] shift_patterns テーブル
- constraint_violations JSONB, score DECIMAL(5,2) 一致

### [一致] shift_entries テーブル
- is_manual_edit BOOLEAN 一致

### [一致] generation_jobs テーブル
- status: pending/processing/completed/failed 一致

### [一致] インデックス
- 全インデックス定義がマイグレーションに含まれている

---

## 3. 画面照合（docs/SCREENS.md vs 実装）

### [一致] ダッシュボード (/)
- スタッフ数、希望提出率、シフト状態、制約数の4カード表示
- ミニカレンダー（日ごと出勤予定人数）ヒートマップ実装済み
- shift_status のラベル定義一致

### [対応: 修正済み] ダッシュボード - PDF出力ボタン
- **設計書**: クイックアクションに「PDFを出力」ボタンあり
- **実装**: 「シフト希望を入力」「シフトを生成」の2つのみ
- **対応内容**: シフト生成済み状態の場合に「PDFを出力」ボタンを追加（`Dashboard/index.tsx`）

### [一致] スタッフ管理 (/staffs)
- テーブル: 名前、役割、雇用形態、状態、操作
- 新規追加/編集モーダル: 名前、役割(select)、雇用形態(select)
- SCREENS.md のワイヤーフレームに「月間上限H」列は存在せず、実装にも含まれていない（正しく除去済み）

### [一致] シフト希望入力 (/requests)
- スタッフセレクター
- 月間希望時間: min 〜 max のレンジ入力対応済み
- カレンダー形式: ○/x/△ 切替、ダブルクリックで時間帯入力

### [一致] 制約設定 (/constraints)
- ハード/ソフト制約分類表示
- 制約追加モーダル: 名前、種類、カテゴリ、カテゴリ固有設定
- 優先度（ソフト制約のみ）

### [一致] シフト生成 (/generate)
- 生成条件表示、パターン数選択、プログレスバー
- パターンカード: スコア、違反数、詳細/選択ボタン

### [一致] シフト編集 (/shifts/:patternId/edit)
- 週ナビゲーション、セル編集（ポップオーバー）
- PDF 出力・確定ボタン
- サマリーテーブル（名前、月間H、差分）
- 手動編集セルの色分け表示

---

## 4. LLM 連携照合（docs/LLM_INTEGRATION.md vs 実装）

### [一致] システムプロンプト
- `generator.go:121-152` の内容が LLM_INTEGRATION.md のシステムプロンプトと完全一致

### [一致] ユーザープロンプト構造
- スタッフ情報、月間設定、シフト希望、ハード/ソフト制約、追加指示の構造一致
- patternIdx > 0 の場合に「異なるアプローチ」指示を追加

### [一致] constraint_violations フォーマット
- `constraint_name` + `type` + `message` の3フィールド統一
- model.ConstraintViolation、LLMResponse、ShiftPattern で統一

### [一致] バリデーション→再生成フロー
- `shift_service.go:78-162` で maxRetries = 3 のリトライループ実装
- ハード制約違反時のリトライ、フロー一致

### [一致] API 設定
- モデル: `claude-sonnet-4-5-20250929` 一致
- MaxTokens: 8192 一致
- Temperature: 0.7 一致

---

## 5. フロント API クライアント照合（docs/API.md vs frontend/src/api/）

### [対応: 修正済み] monthlySettingsApi に batchCreate 関数が欠落
- **設計書**: `POST /api/v1/staff-monthly-settings/batch` が定義
- **実装**: `monthlySettingsApi` に `batchCreate` メソッドがなかった
- **対応内容**: `monthlySettingsApi.ts` に `batchCreate` メソッドを追加

### [対応: 修正済み] GenerationJob 型の id フィールド不一致
- **設計書**: バックエンド GenerationJob モデルは `id` フィールド
- **実装**: フロントエンド型定義で `job_id` になっていた
- **対応内容**: `types/index.ts` の GenerationJob.job_id を `id` に修正、shiftStore.ts の参照も修正

### [対応: 修正済み] ShiftEntry 型に pattern_id が欠落
- **設計書**: API.md のエントリレスポンスに `pattern_id` フィールドあり
- **実装**: フロントエンド ShiftEntry 型に `pattern_id`、`created_at`、`updated_at` が未定義
- **対応内容**: `types/index.ts` に `pattern_id`、`created_at`、`updated_at` を追加

### [対応: 修正済み] rest_hours 制約の config キー不一致
- **設計書**: DATABASE.md で `{ "min_hours": 11 }`
- **実装**: Constraints UI で `min_rest_hours` を使用
- **対応内容**: `Constraints/index.tsx` の config キーを `min_hours` に統一

### [対応: 修正済み] staff_compatibility 制約の config フォーマット不一致
- **設計書**: DATABASE.md で `{ "staff_ids": ["uuid-1", "uuid-2"], "rule": "prefer_together" }`
- **実装**: UI で `staff_id_1`, `staff_id_2` の別フィールドを使用、rule フィールドなし
- **対応内容**: `Constraints/index.tsx` を `staff_ids` 配列 + `rule` セレクターに修正

### [一致] staffApi
- list, get, create, update, delete 全メソッド対応済み

### [一致] shiftRequestApi
- list, create, batchCreate, update, delete 全メソッド対応済み

### [一致] constraintApi
- list, create, update, delete 全メソッド対応済み

### [一致] shiftApi
- generate, getJobStatus, listPatterns, getPattern, selectPattern, finalizePattern, createEntry, updateEntry, deleteEntry 全メソッド対応済み

### [一致] dashboardApi
- getSummary メソッド対応済み

---

## 設計変更推奨

### [対応: 設計変更推奨] バリデーター - 勤務間インターバルチェック未実装
- **設計書**: LLM_INTEGRATION.md バリデーションチェック項目 #4 に「勤務間インターバル」チェックあり
- **実装**: `shift_validator.go` に rest_hours カテゴリの制約チェックが未実装
- **推奨**: バリデーターに `checkRestHours` メソッドを追加し、前日終業〜翌日始業の間隔を検証するロジックを実装すべき。現時点では LLM にプロンプトで指示しているが、プログラム的検証が望ましい。

---

## 修正ファイル一覧

| ファイル | 修正内容 |
|---------|---------|
| `frontend/src/api/monthlySettingsApi.ts` | batchCreate メソッド追加 |
| `frontend/src/types/index.ts` | GenerationJob.job_id -> id、ShiftEntry に pattern_id/created_at/updated_at 追加、GenerationJob に created_at 追加 |
| `frontend/src/stores/shiftStore.ts` | GenerationJob 構築時のフィールド名修正 |
| `frontend/src/pages/Dashboard/index.tsx` | PDF出力ボタン追加 |
| `frontend/src/pages/Constraints/index.tsx` | rest_hours config キー修正、staff_compatibility config フォーマット修正 |
