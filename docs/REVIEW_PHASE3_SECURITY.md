# Phase 3 セキュリティ・品質レビュー

## サマリー
- レビューファイル数: 42
- 修正済み問題: 9件
- 警告（修正不要）: 5件

---

## 修正済み問題一覧

### [重要度: HIGH] スタッフ作成時の入力バリデーション不足
- **ファイル**: `backend/internal/service/staff_service.go`
- **問題**: `name` に長さ制限がなく、`role` / `employment_type` に許可値チェックがない。DB の VARCHAR(100) / VARCHAR(50) に対して事前バリデーションがないため、不正な値が登録される可能性がある
- **修正内容**: `name` に100文字制限、`role` に `kitchen/hall/both` の許可値チェック、`employment_type` に `full_time/part_time` の許可値チェックを追加

### [重要度: HIGH] シフト希望の request_type バリデーション不足
- **ファイル**: `backend/internal/service/shift_request_service.go`
- **問題**: `request_type` に任意の文字列が指定可能。`year_month` のフォーマット検証もない
- **修正内容**: `request_type` に `available/unavailable/preferred` の許可値チェック、`year_month` に長さ7(YYYY-MM)のフォーマットチェックを追加

### [重要度: HIGH] 制約条件の type/category バリデーション不足
- **ファイル**: `backend/internal/service/constraint_service.go`
- **問題**: `type` に `hard/soft` 以外の値、`category` に未定義のカテゴリが指定可能。`name` にも長さ制限なし
- **修正内容**: `name` に200文字制限、`type` に `hard/soft` の許可値チェック、`category` に定義済みカテゴリ7種のホワイトリストチェックを追加

### [重要度: HIGH] シフト生成の PatternCount に上限制限なし（リソース枯渇）
- **ファイル**: `backend/internal/service/shift_service.go`
- **問題**: `pattern_count` に上限がなく、大きな数値を指定すると Claude API への大量リクエストが発生し、API コストとサーバーリソースの枯渇につながる
- **修正内容**: `pattern_count` の上限を5に制限。`year_month` にもフォーマットチェックを追加

### [重要度: HIGH] BatchCreate にサイズ制限がない（DoS対策）
- **ファイル**: `backend/internal/service/shift_request_service.go`, `backend/internal/service/staff_monthly_setting_service.go`
- **問題**: バッチ API に件数上限がなく、大量データの一括送信による DoS 攻撃が可能
- **修正内容**: 両方の BatchCreate に100件の上限制限を追加

### [重要度: MEDIUM] CORS 設定がハードコード
- **ファイル**: `backend/internal/middleware/cors.go`
- **問題**: 許可オリジンが `localhost:5173` と `localhost:3000` にハードコードされており、本番デプロイ時に変更が必要
- **修正内容**: 環境変数 `CORS_ALLOWED_ORIGINS` からカンマ区切りでオリジンを設定可能に変更。未設定時はデフォルト値を使用

### [重要度: MEDIUM] 月間設定の時間バリデーション不足
- **ファイル**: `backend/internal/service/staff_monthly_setting_service.go`
- **問題**: `min_preferred_hours` / `max_preferred_hours` に負の値や論理的に不正な値（min > max）が指定可能
- **修正内容**: 0以上チェック、min <= max チェック、上限744時間（31日 x 24時間）チェックを追加

### [重要度: MEDIUM] internalError でサーバーサイドログが出力されない
- **ファイル**: `backend/internal/handler/helper.go`
- **問題**: `internalError` 関数がクライアントにはジェネリックなエラーメッセージを返すが、サーバーログに実際のエラー内容を出力していないため、障害調査ができない
- **修正内容**: `log.Printf` でリクエスト情報とエラー内容をサーバーログに出力するように追加

### [重要度: MEDIUM] BulkCreate がトランザクション外で実行されている
- **ファイル**: `backend/internal/repository/shift_entry_repository.go`
- **問題**: `BulkCreate` がトランザクションなしで複数 INSERT を実行しており、途中失敗時にデータ不整合（一部のみ登録済み）が発生する
- **修正内容**: `tx.Begin()` / `tx.Commit()` でトランザクション内実行に変更。失敗時は `defer tx.Rollback()` で自動ロールバック

---

## 警告（修正不要）

### [重要度: LOW] DashboardService でエラーが無視されている箇所
- **ファイル**: `backend/internal/service/dashboard_service.go:70,83,103,109`
- **問題**: `patterns, _ := s.patternRepo.ListByYearMonth(...)` や `dailyCounts, _ := s.entryRepo.DailyStaffCounts(...)` 等でエラーが `_` で無視されている
- **理由**: ダッシュボードのサマリーデータは「ベストエフォート」の表示であり、部分的なデータ欠落は許容される設計と判断。ただし本番環境ではロギングの追加を推奨

### [重要度: LOW] ShiftGenerate ページのポーリング処理
- **ファイル**: `frontend/src/pages/ShiftGenerate/index.tsx:106`
- **問題**: `setInterval` 内で async 関数を呼んでおり、前のポーリングが完了する前に次のポーリングが開始される可能性がある
- **理由**: ポーリング間隔が2秒と十分に長く、通常の API レスポンスタイムでは問題にならない。長期的には `setTimeout` ベースに変更することを推奨

### [重要度: LOW] staffName 取得時のエラー無視
- **ファイル**: `backend/internal/repository/shift_request_repository.go:80`, `backend/internal/repository/shift_entry_repository.go:73`, `backend/internal/repository/staff_monthly_setting_repository.go:86`
- **問題**: RETURNING 後に `staff_name` を別クエリで取得する際、エラーを `_` で無視している
- **理由**: JOIN で取得するのが理想的だが、INSERT/UPDATE の RETURNING に JOIN は使えないため、表示名が空になるだけで機能上の問題はない

### [重要度: LOW] ShiftEdit ページの editPopover の位置
- **ファイル**: `frontend/src/pages/ShiftEdit/index.tsx:230`
- **問題**: 編集ポップオーバーが `position: absolute` で表示されるが、クリックされたセルとの相対位置が指定されていない
- **理由**: UI/UX の問題であり、セキュリティや品質に直接影響しない

### [重要度: INFO] N+1 クエリのリスク
- **ファイル**: `backend/internal/service/shift_service.go:175-176`
- **問題**: `ListPatterns` で各パターンに対して `computeSummary` を呼び出し、内部で `ListByPatternID` を実行している。パターン数 x エントリ取得のN+1クエリ
- **理由**: パターン数は最大5件であり、現実的なパフォーマンス問題にはならない。スケーリング時には一括取得に変更を推奨

---

## セキュリティ良好な点

1. **SQLインジェクション対策**: 全SQLクエリでパラメータバインド(`$1`, `$2`, ...)を使用しており、文字列結合によるSQL組み立ては行っていない
2. **XSS対策**: React のデフォルトエスケープに依拠しており、`dangerouslySetInnerHTML` の使用なし
3. **環境変数管理**: APIキーは環境変数で管理され、ハードコードなし。`.env.example` のみ存在
4. **パストラバーサル**: ファイルパス操作なし
5. **アーキテクチャ**: handler -> service -> repository の3層構造が守られており、handler が直接 DB を触っていない
6. **API呼び出し管理**: フロントエンドの API 呼び出しは全て `api/` ディレクトリ + zustand store 経由で管理されている
7. **goroutine管理**: `runGeneration` は `context.Background()` を使用しており、リクエストコンテキストのキャンセルに依存しない適切な設計
