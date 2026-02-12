# シフト作成アプリ - 最終サマリー

## プロジェクト概要

飲食店・クリーニング店舗向けのシフト作成アプリケーション。
管理者が従業員のシフト希望と制約条件を入力し、LLM（Claude API）が最適なシフトパターンを複数生成する。

- **ユーザー**: 管理者のみ（従業員数 5〜10人）
- **技術スタック**: React + Go + PostgreSQL + Claude API
- **デプロイ**: ローカル（Docker Compose）、将来的に AWS

## ワークフロー

```
1. スタッフ登録
2. 月間希望時間の設定（従業員ごと、min〜max）
3. シフト希望の日別入力（出勤可/不可/希望）
4. 制約条件の設定（ハード制約/ソフト制約）
5. LLM によるシフトパターン自動生成（複数パターン）
6. パターン比較・選択
7. 手動編集（微調整）
8. PDF 出力
```

## 開発プロセス

### チーム構成（延べ7エージェント）

| Agent | ロール | フェーズ |
|-------|--------|---------|
| PjM（Team Lead） | タスク管理・統括・統合修正 | 全フェーズ |
| pdm | 設計品質管理・設計書照合 | Phase 1, 3 |
| infra-dev | インフラ・環境構築 | Phase 1 |
| backend-dev | Go バックエンド開発 | Phase 2 |
| frontend-dev | React フロントエンド開発 | Phase 2 |
| veteran-eng | セキュリティ・品質レビュー | Phase 3 |
| tester | テスト作成・実行 | Phase 3 |

### Phase 1: 設計・環境構築

| タスク | 担当 | 結果 |
|--------|------|------|
| プロジェクト初期化 | infra-dev | Go mod + Vite React TS 雛形作成 |
| Docker Compose 作成 | infra-dev | Go(8080) + React(5173) + PostgreSQL(5432) |
| DB マイグレーション | infra-dev + PjM | 7テーブル + インデックス |
| 設計書整合性チェック | pdm + PjM | 7件の不整合を検出・修正 |

### Phase 2: 実装

| タスク | 担当 | 結果 |
|--------|------|------|
| バックエンド API 全実装 | backend-dev | 25ファイル、全エンドポイント |
| フロントエンド UI 全実装 | frontend-dev | 6画面 + PDF出力、ビルド成功 |

### Phase 3: レビュー・テスト

| タスク | 担当 | 結果 |
|--------|------|------|
| セキュリティ・品質レビュー | veteran-eng | 9件修正（入力バリデーション強化、CORS改善等） |
| ユニットテスト | tester + PjM | 38テスト / 90サブテスト ALL PASS |
| 設計書 vs 実装照合 | pdm | 28項目中22一致、6件修正 |

### 残課題修正（PjM が対応）

| 課題 | 対応 |
|------|------|
| 勤務間インターバル（rest_hours）バリデータ未実装 | `shift_validator.go` に `checkRestHours` を追加、テスト6件追加・全パス |

## アーキテクチャ

```
Frontend (React)        Backend (Go)           Database (PostgreSQL)
┌────────────────┐      ┌────────────────┐     ┌────────────────┐
│ React 18 + TS  │─REST─│ Echo v4        │─SQL─│ PostgreSQL 16  │
│ Zustand        │      │ pgx/v5         │     │ 7 tables       │
│ FullCalendar   │      │ Anthropic SDK  │     │ JSONB configs  │
│ jsPDF          │      │ Validator      │     └────────────────┘
└────────────────┘      └───────┬────────┘
                                │ HTTPS
                        ┌───────▼────────┐
                        │ Claude API     │
                        │ (Sonnet 4.5)   │
                        └────────────────┘
```

## データベース（7テーブル）

| テーブル | 説明 |
|---------|------|
| staffs | スタッフ情報 |
| staff_monthly_settings | 月間希望労働時間（min〜max） |
| shift_requests | 日別シフト希望 |
| constraints | 制約条件（JSONB config） |
| shift_patterns | LLM生成パターン |
| shift_entries | パターン内の個別シフト |
| generation_jobs | 非同期生成ジョブ管理 |

## API エンドポイント

| カテゴリ | エンドポイント数 |
|---------|----------------|
| スタッフ管理 | 4 (CRUD) |
| スタッフ月間設定 | 5 (CRUD + batch) |
| シフト希望 | 5 (CRUD + batch) |
| 制約条件 | 4 (CRUD) |
| ダッシュボード | 1 |
| シフト生成 | 2 (generate + status) |
| シフトパターン | 4 (list/detail/select/finalize) |
| シフトエントリ | 3 (CRUD) |
| **合計** | **28 エンドポイント** |

## 画面構成

| # | 画面 | パス | 主な機能 |
|---|------|------|---------|
| 1 | ダッシュボード | `/` | 月間概要、クイックアクション |
| 2 | スタッフ管理 | `/staffs` | CRUD |
| 3 | シフト希望入力 | `/requests` | カレンダー + 月間設定 |
| 4 | 制約設定 | `/constraints` | ハード/ソフト制約管理 |
| 5 | シフト生成 | `/generate` | LLM生成 + パターン比較 |
| 6 | シフト編集 | `/shifts/:id/edit` | グリッド編集 + PDF出力 |

## バリデーション（ハイブリッド方式）

LLM生成 → プログラム的バリデーション → 違反時は再生成（最大3回）

### ハード制約チェック
- 出勤不可日チェック
- 時間整合性（start < end、対象月内）
- 重複チェック（同一スタッフ同日）
- 連勤制限（max_consecutive_days）
- 最低/最大スタッフ数（min_staff / max_staff）
- 勤務間インターバル（rest_hours）

### ソフト制約チェック
- 月間労働時間の希望との乖離
- スタッフ相性（staff_compatibility）

## テスト結果

```
go test ./... -v -count=1

shift-app/internal/handler   ... PASS (12 tests)
shift-app/internal/model     ... PASS (1 test / 4 subtests)
shift-app/internal/service   ... PASS (17 tests)
shift-app/internal/validator ... PASS (8 tests / 46 subtests)

Total: 38 test functions / 90 subtests - ALL PASS
```

## セキュリティレビュー結果

- SQLインジェクション: 全クエリでパラメータバインド使用 ✓
- XSS: dangerouslySetInnerHTML 未使用 ✓
- 入力バリデーション: 全エンドポイントで実施 ✓
- CORS: 環境変数で制御可能 ✓
- 環境変数: APIキーはハードコードなし ✓
- バッチ処理: サイズ制限あり（100件） ✓
- トランザクション: バルク操作はトランザクション内実行 ✓

## 起動方法

```bash
# .env に ANTHROPIC_API_KEY を設定

# 起動
docker-compose up --build

# アクセス
# フロント: http://localhost:5173
# API:     http://localhost:8080/api/v1/
```

## 設計ドキュメント一覧

| ファイル | 内容 |
|---------|------|
| docs/ARCHITECTURE.md | 全体アーキテクチャ |
| docs/DATABASE.md | DB スキーマ |
| docs/API.md | API 仕様（28エンドポイント） |
| docs/SCREENS.md | 画面設計 |
| docs/LLM_INTEGRATION.md | LLM連携設計 |
| docs/TEAM_PLAN.md | チーム計画 |
| docs/REVIEW_PHASE1.md | Phase 1 設計レビュー |
| docs/REVIEW_PHASE3_SECURITY.md | Phase 3 セキュリティレビュー |
| docs/REVIEW_PHASE3_TEST.md | Phase 3 テストレポート |
| docs/REVIEW_PHASE3_CONFORMANCE.md | Phase 3 設計照合レポート |
| docs/FINAL_SUMMARY.md | 本ドキュメント |
