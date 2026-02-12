# チーム計画書

## チーム構成

| # | Agent名 | ロール | 稼働フェーズ | subagent_type |
|---|---------|--------|------------|---------------|
| 1 | pjm | PjM（Team Lead） | 全フェーズ | team-lead |
| 2 | pdm | PdM（設計品質管理） | Phase 1, 3 | general-purpose |
| 3 | infra-dev | インフラ・環境構築 | Phase 1 | general-purpose |
| 4 | backend-dev | Go バックエンド開発 | Phase 2 | general-purpose |
| 5 | frontend-dev | React フロントエンド開発 | Phase 2 | general-purpose |
| 6 | veteran-eng | ベテランエンジニア（レビュー） | Phase 3 | general-purpose |
| 7 | tester | テスト実行者 | Phase 3 | general-purpose |

## フェーズ計画

### Phase 1: 設計・環境構築（同時稼働: PjM + PdM + infra-dev）

**目標:** プロジェクトの土台を完成させる

| タスクID | タスク | 担当 | 依存 | 成果物 |
|---------|--------|------|------|--------|
| T1 | プロジェクト初期化（Go mod, React Vite, ディレクトリ構成） | infra-dev | - | backend/, frontend/ の雛形 |
| T2 | Docker Compose 作成（Go + React + PostgreSQL） | infra-dev | T1 | docker-compose.yml, Dockerfile×2 |
| T3 | DB マイグレーション作成（全テーブル） | infra-dev | T1 | migrations/*.sql |
| T4 | 設計書の整合性チェック（API ↔ DB ↔ 画面） | pdm | - | レビューレポート |
| T5 | タスク分解・依存関係定義・Phase 2 準備 | pjm | - | タスクリスト |

**Phase 1 完了条件:**
- `docker-compose up` で Go + React + PostgreSQL が起動する
- マイグレーションが実行され、全テーブルが作成される
- PdM の設計整合性チェックが完了し、問題が解消されている

---

### Phase 2: 実装（同時稼働: PjM + backend-dev + frontend-dev）

**目標:** バックエンド API とフロントエンド UI を実装する

#### バックエンド（backend-dev）

| タスクID | タスク | 依存 | 成果物 |
|---------|--------|------|--------|
| T6 | Go プロジェクト構造・ルーティング・ミドルウェア設定 | T1,T2 | main.go, router, middleware |
| T7 | スタッフ CRUD API | T6 | handler/staff.go, service, repository |
| T8 | シフト希望 CRUD API | T6 | handler/shift_request.go |
| T9 | 制約条件 CRUD API | T6 | handler/constraint.go |
| T10 | LLM 連携・シフト生成ロジック | T6 | llm/generator.go, service/shift.go |
| T11 | バリデーションロジック | T10 | validator/shift_validator.go |
| T12 | シフトパターン・エントリ API | T10,T11 | handler/shift.go |
| T13 | 非同期ジョブ管理（生成ジョブ） | T10 | service/job.go |

#### フロントエンド（frontend-dev）

| タスクID | タスク | 依存 | 成果物 |
|---------|--------|------|--------|
| T14 | レイアウト・ルーティング・共通コンポーネント | T1 | Layout/, Common/, App.tsx |
| T15 | API クライアント・型定義 | T14 | api/, types/ |
| T16 | Zustand ストア設計 | T14 | stores/ |
| T17 | スタッフ管理画面 | T14,T15,T16 | pages/Staff/ |
| T18 | シフト希望入力画面（カレンダー） | T14,T15,T16 | pages/ShiftRequest/ |
| T19 | 制約設定画面 | T14,T15,T16 | pages/Constraints/ |
| T20 | シフト生成画面（パターン比較） | T14,T15,T16 | pages/ShiftGenerate/ |
| T21 | シフト編集画面（グリッド編集） | T14,T15,T16 | pages/ShiftEdit/ |
| T22 | ダッシュボード画面 | T17-T21 | pages/Dashboard/ |
| T23 | PDF出力機能 | T21 | PDF/ShiftPDFExporter.tsx |

**Phase 2 完了条件:**
- 全 API エンドポイントが動作する
- 全画面が表示され、API と連携して動作する
- シフト生成〜編集〜PDF出力のワークフローが一通り動作する

---

### Phase 3: レビュー・テスト（同時稼働: PjM + veteran-eng + tester）

**目標:** 品質保証とバグ修正

#### ベテランエンジニア（veteran-eng）

| タスクID | タスク | 依存 | 成果物 |
|---------|--------|------|--------|
| T24 | セキュリティレビュー（SQLインジェクション、XSS等） | T7-T13 | レビュー指摘リスト |
| T25 | コード品質レビュー（Go/React ベストプラクティス） | T7-T23 | レビュー指摘リスト |
| T26 | アーキテクチャ整合性チェック | T7-T23 | レビュー指摘リスト |
| T27 | レビュー指摘の修正実施 | T24-T26 | 修正コミット |

#### テスト実行者（tester）

| タスクID | タスク | 依存 | 成果物 |
|---------|--------|------|--------|
| T28 | バックエンド ユニットテスト（handler, service） | T7-T13 | *_test.go |
| T29 | バリデーションロジック テスト | T11 | validator/*_test.go |
| T30 | フロントエンド コンポーネントテスト | T14-T23 | *.test.tsx |
| T31 | 統合テスト（API ↔ フロント） | T28-T30 | 統合テストスイート |

#### PdM（Phase 3 で再稼働）

| タスクID | タスク | 依存 | 成果物 |
|---------|--------|------|--------|
| T32 | 設計書 vs 実装コードの全件照合 | T7-T23 | 差分レポート |
| T33 | 差分の修正指示・確認 | T32 | 修正完了確認 |

**Phase 3 完了条件:**
- セキュリティ・品質レビュー指摘が全て解消
- ユニットテストが全てパス
- 設計書と実装の差分が解消
- ワークフロー全体が正常に動作する

---

## 依存関係グラフ

```
Phase 1:
  T1 ──→ T2 ──→ T3
  T4 (並行)
  T5 (並行)

Phase 2:
  T6 ──→ T7, T8, T9 (並行)
  T6 ──→ T10 ──→ T11 ──→ T12
  T10 ──→ T13

  T14 ──→ T15, T16 (並行)
  T15, T16 ──→ T17, T18, T19, T20, T21 (並行)
  T17-T21 ──→ T22
  T21 ──→ T23

Phase 3:
  T7-T23 ──→ T24, T25, T26, T28, T29, T30, T32 (並行)
  T24-T26 ──→ T27
  T28-T30 ──→ T31
  T32 ──→ T33
```

## リスク・注意点

1. **LLM 出力の不安定性**: バリデーション + リトライで対処するが、テストでは LLM の呼び出しをモックにする
2. **フロント↔バック連携**: Phase 2 の初期で API 仕様を厳密に合わせる（型定義の共有）
3. **FullCalendar のカスタマイズ**: カレンダーライブラリの制約でUIが想定通りにならない場合、グリッド自作にフォールバック
4. **Docker 環境**: M1/M2 Mac での PostgreSQL イメージ互換性に注意
