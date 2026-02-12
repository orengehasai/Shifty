package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

const (
	defaultModel     = "claude-sonnet-4-5-20250929"
	defaultMaxTokens = 16384
)

type Generator struct {
	client *anthropic.Client
	db     *pgxpool.Pool
}

func NewGenerator(apiKey string, db *pgxpool.Pool) *Generator {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &Generator{
		client: &client,
		db:     db,
	}
}

func (g *Generator) Generate(ctx context.Context, yearMonth string, patternCount int, patternIdx int, previousPatterns []model.LLMResponse, lastViolations []model.Violation) (*model.LLMResponse, error) {
	// Collect data
	staffs, err := g.getStaffs(ctx)
	if err != nil {
		return nil, fmt.Errorf("スタッフ取得エラー: %w", err)
	}

	monthlySettings, err := g.getMonthlySettings(ctx, yearMonth)
	if err != nil {
		return nil, fmt.Errorf("月間設定取得エラー: %w", err)
	}

	shiftRequests, err := g.getShiftRequests(ctx, yearMonth)
	if err != nil {
		return nil, fmt.Errorf("シフト希望取得エラー: %w", err)
	}

	constraints, err := g.getConstraints(ctx)
	if err != nil {
		return nil, fmt.Errorf("制約条件取得エラー: %w", err)
	}

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(yearMonth, staffs, monthlySettings, shiftRequests, constraints, patternIdx, previousPatterns, lastViolations)

	message, err := g.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:       defaultModel,
		MaxTokens:   defaultMaxTokens,
		Temperature: anthropic.Float(0.7),
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Claude API エラー: %w", err)
	}

	// Extract text content
	var responseText string
	for _, block := range message.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	// Try to parse JSON from response
	responseText = extractJSON(responseText)

	var result model.LLMResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("レスポンスのJSONパース失敗: %w\nResponse: %s", err, responseText)
	}

	return &result, nil
}

func extractJSON(text string) string {
	// Try to find JSON block in markdown code fences
	if idx := strings.Index(text, "```json"); idx != -1 {
		start := idx + 7
		if end := strings.Index(text[start:], "```"); end != -1 {
			return strings.TrimSpace(text[start : start+end])
		}
	}
	if idx := strings.Index(text, "```"); idx != -1 {
		start := idx + 3
		if end := strings.Index(text[start:], "```"); end != -1 {
			return strings.TrimSpace(text[start : start+end])
		}
	}
	// Try to find raw JSON
	if idx := strings.Index(text, "{"); idx != -1 {
		lastIdx := strings.LastIndex(text, "}")
		if lastIdx > idx {
			return text[idx : lastIdx+1]
		}
	}
	return text
}

func buildSystemPrompt() string {
	return `あなたは飲食店・小売店向けのシフト作成エキスパートです。
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
}`
}

func buildUserPrompt(yearMonth string, staffs []staffInfo, settings []settingInfo, requests []requestInfo, constraints []constraintInfo, patternIdx int, previous []model.LLMResponse, lastViolations []model.Violation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("以下の条件で %s のシフトを作成してください。\n\n", yearMonth))

	sb.WriteString("## 店舗営業情報\n")
	sb.WriteString("- 営業時間: 9:00〜22:00\n")
	sb.WriteString(fmt.Sprintf("- 対象期間: %s の全日\n\n", yearMonth))

	sb.WriteString("## スタッフ情報\n")
	for _, s := range staffs {
		sb.WriteString(fmt.Sprintf("- %s(id: %s): %s, %s\n", s.Name, s.ID, s.Role, s.EmploymentType))
	}
	sb.WriteString("\n")

	sb.WriteString("## 月間労働時間の希望\n")
	for _, s := range settings {
		sb.WriteString(fmt.Sprintf("- %s: %d〜%dh で働きたい", s.StaffName, s.MinHours, s.MaxHours))
		if s.Note != "" {
			sb.WriteString(fmt.Sprintf("（%s）", s.Note))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## シフト希望\n")
	staffRequests := make(map[string][]requestInfo)
	for _, r := range requests {
		staffRequests[r.StaffName] = append(staffRequests[r.StaffName], r)
	}
	for name, reqs := range staffRequests {
		sb.WriteString(fmt.Sprintf("- %s: ", name))
		parts := []string{}
		for _, r := range reqs {
			symbol := "○"
			if r.RequestType == "unavailable" {
				symbol = "×"
			} else if r.RequestType == "preferred" {
				symbol = "◎"
			}
			timeStr := ""
			if r.StartTime != "" && r.EndTime != "" {
				timeStr = fmt.Sprintf("(%s-%s)", r.StartTime, r.EndTime)
			}
			parts = append(parts, fmt.Sprintf("%s %s%s", r.Date, symbol, timeStr))
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	var hardConstraints, softConstraints []constraintInfo
	for _, c := range constraints {
		if c.Type == "hard" {
			hardConstraints = append(hardConstraints, c)
		} else {
			softConstraints = append(softConstraints, c)
		}
	}

	sb.WriteString("## ハード制約（必ず守ること）\n")
	sb.WriteString("- 出勤不可マーク(×)の日は必ず休みにする\n")
	for _, c := range hardConstraints {
		sb.WriteString(fmt.Sprintf("- %s\n", c.Description))
	}
	sb.WriteString("\n")

	sb.WriteString("## ソフト制約（できるだけ守ること、優先度順）\n")
	for _, c := range softConstraints {
		sb.WriteString(fmt.Sprintf("- [P:%d] %s\n", c.Priority, c.Description))
	}
	sb.WriteString("\n")

	if patternIdx > 0 {
		sb.WriteString("## 追加指示\n")
		sb.WriteString("前のパターンとは異なるアプローチで作成してください。\n")
		sb.WriteString("例えば、週末のシフト配分を変える、早番/遅番の割り当てを変える等。\n\n")
	}

	if len(lastViolations) > 0 {
		sb.WriteString("## ⚠️ 前回の生成結果で以下の制約違反が検出されました。必ず修正してください。\n")
		for _, v := range lastViolations {
			detail := v.Message
			if v.Date != "" {
				detail += fmt.Sprintf(" (日付: %s)", v.Date)
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", v.Type, v.Constraint, detail))
		}
		sb.WriteString("\n上記の違反を全て解消した上で、JSON形式で出力してください。")
	} else {
		sb.WriteString("上記のJSON形式で出力してください。")
	}

	return sb.String()
}

type staffInfo struct {
	ID             string
	Name           string
	Role           string
	EmploymentType string
}

type settingInfo struct {
	StaffName string
	MinHours  int
	MaxHours  int
	Note      string
}

type requestInfo struct {
	StaffName   string
	Date        string
	StartTime   string
	EndTime     string
	RequestType string
}

type constraintInfo struct {
	Name        string
	Type        string
	Priority    int
	Description string
}

func (g *Generator) getStaffs(ctx context.Context) ([]staffInfo, error) {
	rows, err := g.db.Query(ctx,
		`SELECT id, name, role, employment_type FROM staffs WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []staffInfo
	for rows.Next() {
		var s staffInfo
		if err := rows.Scan(&s.ID, &s.Name, &s.Role, &s.EmploymentType); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (g *Generator) getMonthlySettings(ctx context.Context, yearMonth string) ([]settingInfo, error) {
	rows, err := g.db.Query(ctx,
		`SELECT s.name, sms.min_preferred_hours, sms.max_preferred_hours, COALESCE(sms.note, '')
		 FROM staff_monthly_settings sms
		 JOIN staffs s ON s.id = sms.staff_id
		 WHERE sms.year_month = $1
		 ORDER BY s.name`, yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []settingInfo
	for rows.Next() {
		var s settingInfo
		if err := rows.Scan(&s.StaffName, &s.MinHours, &s.MaxHours, &s.Note); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (g *Generator) getShiftRequests(ctx context.Context, yearMonth string) ([]requestInfo, error) {
	rows, err := g.db.Query(ctx,
		`SELECT s.name, sr.date::text, COALESCE(sr.start_time::text, ''), COALESCE(sr.end_time::text, ''), sr.request_type
		 FROM shift_requests sr
		 JOIN staffs s ON s.id = sr.staff_id
		 WHERE sr.year_month = $1
		 ORDER BY s.name, sr.date`, yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []requestInfo
	for rows.Next() {
		var r requestInfo
		if err := rows.Scan(&r.StaffName, &r.Date, &r.StartTime, &r.EndTime, &r.RequestType); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (g *Generator) getConstraints(ctx context.Context) ([]constraintInfo, error) {
	rows, err := g.db.Query(ctx,
		`SELECT name, type, COALESCE(priority, 0), config FROM constraints WHERE is_active = true ORDER BY type, priority DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []constraintInfo
	for rows.Next() {
		var c constraintInfo
		var configJSON []byte
		if err := rows.Scan(&c.Name, &c.Type, &c.Priority, &configJSON); err != nil {
			return nil, err
		}
		c.Description = buildConstraintDescription(c.Name, configJSON)
		result = append(result, c)
	}
	return result, rows.Err()
}

func buildConstraintDescription(name string, configJSON []byte) string {
	var config map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return name
	}

	parts := []string{name}
	if maxDays, ok := config["max_days"]; ok {
		parts = append(parts, fmt.Sprintf("(最大%v日)", maxDays))
	}
	if minHours, ok := config["min_hours"]; ok {
		parts = append(parts, fmt.Sprintf("(最低%v時間)", minHours))
	}
	if maxCount, ok := config["max_count"]; ok {
		parts = append(parts, fmt.Sprintf("(最大%v人)", maxCount))
	}

	return strings.Join(parts, " ")
}
