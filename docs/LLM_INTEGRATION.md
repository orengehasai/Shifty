# LLM連携設計書

## 概要

Claude API（Anthropic SDK for Go）を使用して、制約条件に基づくシフトパターンを自動生成する。
ハイブリッド方式を採用し、LLM の出力をプログラム的にバリデーションする。

## 使用モデル

- **モデル**: `claude-sonnet-4-5-20250929`
- **選定理由**: コスト/性能バランスが良く、構造化出力（JSON）に強い

## 生成フロー

```
1. データ収集
   ├── スタッフ情報一覧
   ├── シフト希望（対象月分）
   └── 有効な制約条件

2. プロンプト構築
   ├── システムプロンプト（役割・出力フォーマット定義）
   └── ユーザープロンプト（データ + 制約 + 指示）

3. Claude API 呼び出し（パターンごとに1回）
   ├── パターン1生成
   ├── パターン2生成（「パターン1とは異なるアプローチで」と指示）
   └── パターン3生成（「パターン1,2とは異なるアプローチで」と指示）

4. バリデーション
   ├── JSON パース検証
   ├── ハード制約チェック
   ├── ソフト制約チェック（違反をリスト化）
   └── スコア算出

5. 再生成（必要に応じて）
   ├── ハード制約違反がある場合
   ├── 違反内容をフィードバックして再生成指示
   └── 最大3回までリトライ

6. 結果保存
   ├── shift_patterns テーブルに保存
   └── shift_entries テーブルに保存
```

## プロンプト設計

### システムプロンプト

```
あなたは飲食店・小売店向けのシフト作成エキスパートです。
与えられたスタッフ情報、シフト希望、制約条件に基づいて、
最適なシフトスケジュールを作成してください。

## 出力ルール
- 必ず指定されたJSON形式で出力してください
- 全ての日付について、各スタッフの勤務/休みを決定してください
- ハード制約は必ず遵守してください
- ソフト制約はできる限り尊重し、守れない場合は理由を説明してください
- 6時間以上の勤務には60分の休憩を自動付与してください
- スタッフの月間労働時間が希望に近づくよう調整してください

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

### ユーザープロンプト（テンプレート）

```
以下の条件で {year_month} のシフトを作成してください。

## 店舗営業情報
- 営業時間: {operating_hours}
- 対象期間: {year_month} の全日

## スタッフ情報
{staff_list}
（例:
- 田中太郎(id: xxx): キッチン, 正社員
- 佐藤花子(id: yyy): ホール, パート
）

## 月間労働時間の希望
{monthly_settings}
（例:
- 田中太郎: 100〜120h で働きたい
- 佐藤花子: 60〜80h で働きたい（扶養控除のため上限80h）
）

## シフト希望
{shift_requests}
（例:
- 田中太郎: 3/1 ○(9-17), 3/2 ×, 3/3 ○(9-17), ...
- 佐藤花子: 3/1 ○(10-15), 3/2 ○(10-15), 3/3 ×, ...
）

## ハード制約（必ず守ること）
{hard_constraints}
（例:
- ランチタイム(11:00-14:00)は最低3名必要
- 連勤5日以上禁止
- 勤務間インターバル11時間以上
- 出勤不可マーク(×)の日は必ず休みにする
）

## ソフト制約（できるだけ守ること、優先度順）
{soft_constraints}
（例:
- [P:3] 田中と佐藤をできるだけ同じシフトに
- [P:2] パートスタッフは月80h前後に
）

## 追加指示
{additional_instruction}
（パターン2,3の場合:「前のパターンとは異なるアプローチで作成してください。
例えば、週末のシフト配分を変える、早番/遅番の割り当てを変える等。」）

上記のJSON形式で出力してください。
```

## バリデーションロジック

### ハード制約チェック（違反時は再生成）

```go
type ValidationResult struct {
    IsValid    bool
    Violations []Violation
    Warnings   []Warning
    Score      float64
}

type Violation struct {
    Type       string // "hard" or "soft"
    Constraint string // 制約名
    Date       string // 違反日（該当する場合）
    StaffID    string // 該当スタッフ（該当する場合）
    Message    string
}
```

**チェック項目:**

| # | チェック | 種類 | 説明 |
|---|---------|------|------|
| 1 | 出勤不可日チェック | ハード | unavailable の日にシフトが入っていないか |
| 2 | 最低スタッフ数 | ハード | 指定時間帯の出勤人数が最低人数以上か |
| 3 | 連勤チェック | ハード | 連続勤務日数が上限を超えていないか |
| 4 | 勤務間インターバル | ハード | 前日終業〜翌日始業が規定時間以上か |
| 5 | 月間労働時間上限 | ハード | max_monthly_hours を超えていないか |
| 6 | 時間整合性 | ハード | start_time < end_time、日付が対象月内か |
| 7 | 重複チェック | ハード | 同一スタッフの同日重複シフトがないか |

### ソフト制約チェック（警告として記録）

| # | チェック | 説明 |
|---|---------|------|
| 1 | 月間希望時間との乖離 | 希望時間との差が大きい場合に警告 |
| 2 | スタッフ相性 | prefer_together / avoid_together の遵守率 |
| 3 | 希望シフト反映率 | preferred の日がどれだけ反映されたか |

### スコア算出

```
スコア = 100 - (ソフト制約違反のペナルティ合計)

ペナルティ計算:
- 月間時間乖離: |希望 - 実績| / 希望 × 優先度 × 5
- 相性違反: 違反回数 × 優先度 × 3
- 希望未反映: 未反映率 × 優先度 × 2
```

## 再生成フロー

```go
func (s *ShiftGenerator) Generate(ctx context.Context, req GenerateRequest) error {
    for patternIdx := 0; patternIdx < req.PatternCount; patternIdx++ {
        var result *LLMResponse

        for retry := 0; retry < maxRetries; retry++ {
            // LLM 呼び出し
            result, err = s.callClaude(ctx, prompt)

            // バリデーション
            validation := s.validator.Validate(result)

            if validation.IsValid || !validation.HasHardViolations() {
                // ハード制約違反なし → 保存して次のパターンへ
                break
            }

            // ハード制約違反あり → 違反内容をフィードバックして再生成
            prompt = s.buildRetryPrompt(result, validation.Violations)
        }

        // DB保存
        s.savePattern(ctx, result, validation)
    }
}
```

## API 設定

```go
type LLMConfig struct {
    APIKey       string
    Model        string // "claude-sonnet-4-5-20250929"
    MaxTokens    int    // 8192（月間シフト全量を出力するため大きめ）
    Temperature  float64 // 0.7（多様性のため少し高め）
    MaxRetries   int    // 3
}
```

## コスト見積もり

10名・1ヶ月（30日）のシフト生成の場合:
- 入力トークン: 約2,000〜3,000（プロンプト）
- 出力トークン: 約3,000〜5,000（JSON + reasoning）
- 3パターン × (1〜3回) = 3〜9回の API 呼び出し
- Sonnet 4.5: 入力$3/MTok, 出力$15/MTok
- 1回の生成コスト: 約$0.05〜$0.08
- 合計: 約$0.15〜$0.72（約20〜100円）
