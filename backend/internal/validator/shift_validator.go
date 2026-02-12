package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

type ShiftValidator struct {
	db *pgxpool.Pool
}

func NewShiftValidator(db *pgxpool.Pool) *ShiftValidator {
	return &ShiftValidator{db: db}
}

func (v *ShiftValidator) Validate(ctx context.Context, yearMonth string, response *model.LLMResponse) (*model.ValidationResult, error) {
	result := &model.ValidationResult{
		IsValid:    true,
		Violations: []model.Violation{},
		Warnings:   []model.Warning{},
		Score:      100.0,
	}

	// Load reference data
	unavailableDates, err := v.getUnavailableDates(ctx, yearMonth)
	if err != nil {
		return nil, err
	}

	constraints, err := v.getActiveConstraints(ctx)
	if err != nil {
		return nil, err
	}

	monthlySettings, err := v.getMonthlySettings(ctx, yearMonth)
	if err != nil {
		return nil, err
	}

	// 1. Check unavailable dates (hard)
	for _, entry := range response.Entries {
		key := entry.StaffID + ":" + entry.Date
		if unavailableDates[key] {
			result.Violations = append(result.Violations, model.Violation{
				Type:       "hard",
				Constraint: "出勤不可日チェック",
				Date:       entry.Date,
				StaffID:    entry.StaffID,
				Message:    fmt.Sprintf("出勤不可の日(%s)にシフトが割り当てられています", entry.Date),
			})
			result.IsValid = false
		}
	}

	// 2. Check time consistency (hard)
	for _, entry := range response.Entries {
		if !isValidTimeRange(entry.StartTime, entry.EndTime) {
			result.Violations = append(result.Violations, model.Violation{
				Type:       "hard",
				Constraint: "時間整合性",
				Date:       entry.Date,
				StaffID:    entry.StaffID,
				Message:    fmt.Sprintf("開始時刻(%s)が終了時刻(%s)以降です", entry.StartTime, entry.EndTime),
			})
			result.IsValid = false
		}

		// Check date is within target month
		if !strings.HasPrefix(entry.Date, yearMonth) {
			result.Violations = append(result.Violations, model.Violation{
				Type:       "hard",
				Constraint: "日付範囲",
				Date:       entry.Date,
				StaffID:    entry.StaffID,
				Message:    fmt.Sprintf("日付(%s)が対象月(%s)の範囲外です", entry.Date, yearMonth),
			})
			result.IsValid = false
		}
	}

	// 3. Check duplicate shifts (hard)
	seen := make(map[string]bool)
	for _, entry := range response.Entries {
		key := entry.StaffID + ":" + entry.Date
		if seen[key] {
			result.Violations = append(result.Violations, model.Violation{
				Type:       "hard",
				Constraint: "重複チェック",
				Date:       entry.Date,
				StaffID:    entry.StaffID,
				Message:    "同一スタッフの同日に複数のシフトが割り当てられています",
			})
			result.IsValid = false
		}
		seen[key] = true
	}

	// 4. Check constraints
	for _, c := range constraints {
		var config map[string]interface{}
		if err := json.Unmarshal(c.Config, &config); err != nil {
			continue
		}

		switch c.Category {
		case "max_consecutive_days":
			v.checkConsecutiveDays(response.Entries, config, c, result)
		case "min_staff":
			v.checkMinStaff(response.Entries, config, c, result)
		case "max_staff":
			v.checkMaxStaff(response.Entries, config, c, result)
		case "rest_hours":
			v.checkRestHours(response.Entries, config, c, result)
		}
	}

	// 5. Check monthly hours (soft constraints / scoring)
	staffHours := computeStaffHours(response.Entries)
	penalty := 0.0
	for staffID, hours := range staffHours {
		if setting, ok := monthlySettings[staffID]; ok {
			if hours > float64(setting.MaxHours) {
				diff := hours - float64(setting.MaxHours)
				p := diff / float64(setting.MaxHours) * 5.0
				penalty += p
				result.Warnings = append(result.Warnings, model.Warning{
					Type:       "soft_constraint",
					Constraint: "月間労働時間",
					Message:    fmt.Sprintf("月間労働時間(%.0fh)が上限(%dh)を超えています", hours, setting.MaxHours),
				})
			} else if hours < float64(setting.MinHours) {
				diff := float64(setting.MinHours) - hours
				p := diff / float64(setting.MinHours) * 5.0
				penalty += p
				result.Warnings = append(result.Warnings, model.Warning{
					Type:       "soft_constraint",
					Constraint: "月間労働時間",
					Message:    fmt.Sprintf("月間労働時間(%.0fh)が下限(%dh)を下回っています", hours, setting.MinHours),
				})
			}
		}
	}

	result.Score = math.Max(0, 100.0-penalty)

	if len(result.Violations) > 0 {
		result.IsValid = false
	}

	return result, nil
}

func (v *ShiftValidator) checkConsecutiveDays(entries []model.LLMShiftEntry, config map[string]interface{}, c constraintData, result *model.ValidationResult) {
	maxDays := 5
	if md, ok := config["max_days"]; ok {
		if mdFloat, ok := md.(float64); ok {
			maxDays = int(mdFloat)
		}
	}

	// Group dates by staff
	staffDates := make(map[string][]string)
	for _, e := range entries {
		staffDates[e.StaffID] = append(staffDates[e.StaffID], e.Date)
	}

	for staffID, dates := range staffDates {
		sort.Strings(dates)
		consecutive := 1
		for i := 1; i < len(dates); i++ {
			if isConsecutiveDate(dates[i-1], dates[i]) {
				consecutive++
				if consecutive > maxDays {
					result.Violations = append(result.Violations, model.Violation{
						Type:       c.Type,
						Constraint: c.Name,
						StaffID:    staffID,
						Date:       dates[i],
						Message:    fmt.Sprintf("連続勤務%d日（上限%d日）", consecutive, maxDays),
					})
					if c.Type == "hard" {
						result.IsValid = false
					}
				}
			} else {
				consecutive = 1
			}
		}
	}
}

func (v *ShiftValidator) checkMinStaff(entries []model.LLMShiftEntry, config map[string]interface{}, c constraintData, result *model.ValidationResult) {
	// Count staff per date
	dateCounts := make(map[string]int)
	for _, e := range entries {
		dateCounts[e.Date]++
	}

	minCount := 2
	if mc, ok := config["min_count"]; ok {
		if mcFloat, ok := mc.(float64); ok {
			minCount = int(mcFloat)
		}
	}

	for date, count := range dateCounts {
		if count < minCount {
			result.Violations = append(result.Violations, model.Violation{
				Type:       c.Type,
				Constraint: c.Name,
				Date:       date,
				Message:    fmt.Sprintf("%sのスタッフ数(%d)が最低人数(%d)未満", date, count, minCount),
			})
			if c.Type == "hard" {
				result.IsValid = false
			}
		}
	}
}

func (v *ShiftValidator) checkMaxStaff(entries []model.LLMShiftEntry, config map[string]interface{}, c constraintData, result *model.ValidationResult) {
	dateCounts := make(map[string]int)
	for _, e := range entries {
		dateCounts[e.Date]++
	}

	maxCount := 5
	if mc, ok := config["max_count"]; ok {
		if mcFloat, ok := mc.(float64); ok {
			maxCount = int(mcFloat)
		}
	}

	for date, count := range dateCounts {
		if count > maxCount {
			result.Violations = append(result.Violations, model.Violation{
				Type:       c.Type,
				Constraint: c.Name,
				Date:       date,
				Message:    fmt.Sprintf("%sのスタッフ数(%d)が最大人数(%d)を超過", date, count, maxCount),
			})
			if c.Type == "hard" {
				result.IsValid = false
			}
		}
	}
}

func (v *ShiftValidator) checkRestHours(entries []model.LLMShiftEntry, config map[string]interface{}, c constraintData, result *model.ValidationResult) {
	minRestHours := 11.0
	if mh, ok := config["min_hours"]; ok {
		if mhFloat, ok := mh.(float64); ok {
			minRestHours = mhFloat
		}
	}

	// Group entries by staff, sorted by date
	staffEntries := make(map[string][]model.LLMShiftEntry)
	for _, e := range entries {
		staffEntries[e.StaffID] = append(staffEntries[e.StaffID], e)
	}

	for staffID, sEntries := range staffEntries {
		sort.Slice(sEntries, func(i, j int) bool {
			return sEntries[i].Date < sEntries[j].Date
		})

		for i := 1; i < len(sEntries); i++ {
			prev := sEntries[i-1]
			curr := sEntries[i]

			// Only check consecutive dates
			if !isConsecutiveDate(prev.Date, curr.Date) {
				continue
			}

			// Calculate interval: prev end_time to curr start_time
			var peh, pem, csh, csm int
			fmt.Sscanf(prev.EndTime, "%d:%d", &peh, &pem)
			fmt.Sscanf(curr.StartTime, "%d:%d", &csh, &csm)

			// Interval in hours: (24 - prev_end) + curr_start
			intervalMin := (24*60 - (peh*60 + pem)) + (csh*60 + csm)
			intervalHours := float64(intervalMin) / 60.0

			if intervalHours < minRestHours {
				result.Violations = append(result.Violations, model.Violation{
					Type:       c.Type,
					Constraint: c.Name,
					StaffID:    staffID,
					Date:       curr.Date,
					Message:    fmt.Sprintf("勤務間インターバル(%.1fh)が最低%.0fh未満です（前日%s終了→当日%s開始）", intervalHours, minRestHours, prev.EndTime, curr.StartTime),
				})
				if c.Type == "hard" {
					result.IsValid = false
				}
			}
		}
	}
}

func computeStaffHours(entries []model.LLMShiftEntry) map[string]float64 {
	hours := make(map[string]float64)
	for _, e := range entries {
		var sh, sm, eh, em int
		fmt.Sscanf(e.StartTime, "%d:%d", &sh, &sm)
		fmt.Sscanf(e.EndTime, "%d:%d", &eh, &em)
		totalMin := (eh*60 + em) - (sh*60 + sm) - e.BreakMinutes
		if totalMin < 0 {
			totalMin = 0
		}
		hours[e.StaffID] += float64(totalMin) / 60.0
	}
	return hours
}

func isValidTimeRange(start, end string) bool {
	return start < end
}

func isConsecutiveDate(d1, d2 string) bool {
	// Simple check: parse YYYY-MM-DD and check if d2 = d1 + 1 day
	var y1, m1, day1, y2, m2, day2 int
	fmt.Sscanf(d1, "%d-%d-%d", &y1, &m1, &day1)
	fmt.Sscanf(d2, "%d-%d-%d", &y2, &m2, &day2)

	// Same month
	if y1 == y2 && m1 == m2 && day2 == day1+1 {
		return true
	}
	// Cross month boundary
	if y1 == y2 && m2 == m1+1 && day2 == 1 {
		daysInMonth := daysInMonthForYear(y1, m1)
		if day1 == daysInMonth {
			return true
		}
	}
	return false
}

func daysInMonthForYear(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
			return 29
		}
		return 28
	}
	return 30
}

// Internal types
type constraintData struct {
	Name     string
	Type     string
	Category string
	Config   json.RawMessage
}

type monthlySetting struct {
	MinHours int
	MaxHours int
}

func (v *ShiftValidator) getUnavailableDates(ctx context.Context, yearMonth string) (map[string]bool, error) {
	rows, err := v.db.Query(ctx,
		`SELECT staff_id, date::text FROM shift_requests WHERE year_month = $1 AND request_type = 'unavailable'`, yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var staffID, date string
		if err := rows.Scan(&staffID, &date); err != nil {
			return nil, err
		}
		result[staffID+":"+date] = true
	}
	return result, rows.Err()
}

func (v *ShiftValidator) getActiveConstraints(ctx context.Context) ([]constraintData, error) {
	rows, err := v.db.Query(ctx,
		`SELECT name, type, category, config FROM constraints WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []constraintData
	for rows.Next() {
		var c constraintData
		if err := rows.Scan(&c.Name, &c.Type, &c.Category, &c.Config); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (v *ShiftValidator) getMonthlySettings(ctx context.Context, yearMonth string) (map[string]monthlySetting, error) {
	rows, err := v.db.Query(ctx,
		`SELECT staff_id, min_preferred_hours, max_preferred_hours FROM staff_monthly_settings WHERE year_month = $1`, yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]monthlySetting)
	for rows.Next() {
		var staffID string
		var s monthlySetting
		if err := rows.Scan(&staffID, &s.MinHours, &s.MaxHours); err != nil {
			return nil, err
		}
		result[staffID] = s
	}
	return result, rows.Err()
}
