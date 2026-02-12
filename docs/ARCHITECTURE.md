# アーキテクチャ設計書

## 1. システム概要

飲食店・クリーニング店舗向けのシフト作成アプリケーション。
管理者が従業員のシフト希望と制約条件を入力し、LLM（Claude API）が最適なシフトパターンを複数生成する。

### ユーザー
- **管理者のみ**が操作（従業員は直接アクセスしない）
- 従業員数: 5〜10人規模

### ワークフロー
1. 管理者がスタッフ情報を登録
2. 従業員から受け取ったシフト希望を管理者が入力
3. 制約条件を設定（ハード制約・ソフト制約）
4. LLM が制約に基づき複数パターンのシフト案を自動生成
5. 管理者がパターンを比較・選択
6. 選択したパターンを手動編集（微調整）
7. 確定したシフトを PDF 出力

## 2. 技術スタック

| レイヤー | 技術 | バージョン |
|---------|------|-----------|
| フロントエンド | React + TypeScript | React 18+ |
| 状態管理 | Zustand | latest |
| カレンダーUI | FullCalendar | v6 |
| PDF生成 | jsPDF + jspdf-autotable | latest |
| バックエンド | Go (net/http or Echo) | Go 1.22+ |
| データベース | PostgreSQL | 16+ |
| LLM | Claude API (Anthropic SDK) | claude-sonnet-4-5-20250929 |
| コンテナ | Docker + Docker Compose | latest |
| API仕様 | REST (JSON) | - |

## 3. ディレクトリ構成

```
claude-poc/
├── docs/                          # 設計ドキュメント
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go            # エントリーポイント
│   ├── internal/
│   │   ├── config/                # 設定管理
│   │   ├── handler/               # HTTPハンドラー
│   │   ├── service/               # ビジネスロジック
│   │   ├── repository/            # DBアクセス層
│   │   ├── model/                 # データモデル
│   │   ├── llm/                   # LLM連携
│   │   ├── validator/             # シフトバリデーション
│   │   └── middleware/            # ミドルウェア（CORS等）
│   ├── migrations/                # DBマイグレーション
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/            # 共通コンポーネント
│   │   │   ├── Layout/
│   │   │   ├── Calendar/
│   │   │   └── Common/
│   │   ├── pages/                 # ページコンポーネント
│   │   │   ├── Dashboard/
│   │   │   ├── Staff/
│   │   │   ├── ShiftRequest/
│   │   │   ├── Constraints/
│   │   │   ├── ShiftGenerate/
│   │   │   └── ShiftEdit/
│   │   ├── stores/                # Zustand ストア
│   │   ├── api/                   # API クライアント
│   │   ├── types/                 # TypeScript 型定義
│   │   ├── utils/                 # ユーティリティ
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

## 4. 通信フロー

```
┌──────────────┐     HTTP/REST      ┌──────────────┐     SQL      ┌──────────────┐
│   React      │ ◄─────────────────► │   Go API     │ ◄──────────► │  PostgreSQL   │
│   Frontend   │     JSON            │   Server     │              │  Database    │
│   :5173      │                     │   :8080      │              │  :5432       │
└──────────────┘                     └──────┬───────┘              └──────────────┘
                                            │
                                            │ HTTPS (Claude API)
                                            ▼
                                     ┌──────────────┐
                                     │  Anthropic   │
                                     │  Claude API  │
                                     └──────────────┘
```

## 5. シフト生成フロー（非同期）

```
フロント                  バックエンド                    Claude API
  │                          │                              │
  │── POST /api/v1/shifts/generate ──────────►│              │
  │◄── 202 {job_id} ───────────────────────│              │
  │                          │── goroutine開始 ──────────────►│
  │── GET /api/v1/shifts/generate/:job_id ─►│              │
  │◄── {status: processing} ───────────────│              │
  │                          │◄── シフト案(JSON) ────────────│
  │                          │── バリデーション               │
  │                          │── 違反あり → 再生成指示 ──────►│
  │                          │◄── 修正版 ────────────────────│
  │                          │── DB保存                      │
  │── GET /api/v1/shifts/generate/:job_id ─►│              │
  │◄── {status: completed} ────────────────│              │
  │── GET /api/v1/shifts/patterns?year_month=... ──►│      │
  │◄── [pattern1, pattern2..] ─────────────│              │
```

## 6. 将来の拡張性

- **認証追加**: JWT ベースの認証ミドルウェアを追加可能
- **AWS デプロイ**: ECS + RDS + ALB 構成に移行可能
- **通知機能**: LINE / メール通知の追加
- **従業員ポータル**: スタッフ自身がシフト希望を入力する画面
- **制約の拡張**: Constraint テーブルの JSONB カラムで柔軟に対応
