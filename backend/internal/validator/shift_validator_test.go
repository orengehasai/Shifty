package validator

import (
	"encoding/json"
	"testing"

	"shift-app/internal/model"
)

// --- isValidTimeRange tests ---

func TestIsValidTimeRange(t *testing.T) {
	tests := []struct {
		name  string
		start string
		end   string
		want  bool
	}{
		{"normal range", "09:00", "17:00", true},
		{"start equals end", "09:00", "09:00", false},
		{"start after end", "17:00", "09:00", false},
		{"one minute apart", "08:59", "09:00", true},
		{"midnight to morning", "00:00", "06:00", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidTimeRange(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("isValidTimeRange(%q, %q) = %v, want %v", tt.start, tt.end, got, tt.want)
			}
		})
	}
}

// --- isConsecutiveDate tests ---

func TestIsConsecutiveDate(t *testing.T) {
	tests := []struct {
		name string
		d1   string
		d2   string
		want bool
	}{
		{"consecutive same month", "2025-01-15", "2025-01-16", true},
		{"not consecutive same month", "2025-01-15", "2025-01-17", false},
		{"cross month Jan-Feb", "2025-01-31", "2025-02-01", true},
		{"cross month Feb-Mar (non-leap)", "2025-02-28", "2025-03-01", true},
		{"cross month Feb-Mar (leap)", "2024-02-29", "2024-03-01", true},
		{"not cross month", "2025-01-30", "2025-02-01", false},
		{"same day", "2025-01-15", "2025-01-15", false},
		{"cross month Apr-May", "2025-04-30", "2025-05-01", true},
		{"cross month Dec-Jan next year", "2025-12-31", "2026-01-01", false}, // different year handling
		{"reverse order", "2025-01-16", "2025-01-15", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isConsecutiveDate(tt.d1, tt.d2)
			if got != tt.want {
				t.Errorf("isConsecutiveDate(%q, %q) = %v, want %v", tt.d1, tt.d2, got, tt.want)
			}
		})
	}
}

// --- daysInMonthForYear tests ---

func TestDaysInMonthForYear(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month int
		want  int
	}{
		{"January", 2025, 1, 31},
		{"February non-leap", 2025, 2, 28},
		{"February leap (div 4)", 2024, 2, 29},
		{"February leap (div 400)", 2000, 2, 29},
		{"February non-leap (div 100)", 1900, 2, 28},
		{"April", 2025, 4, 30},
		{"June", 2025, 6, 30},
		{"July", 2025, 7, 31},
		{"August", 2025, 8, 31},
		{"September", 2025, 9, 30},
		{"November", 2025, 11, 30},
		{"December", 2025, 12, 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := daysInMonthForYear(tt.year, tt.month)
			if got != tt.want {
				t.Errorf("daysInMonthForYear(%d, %d) = %d, want %d", tt.year, tt.month, got, tt.want)
			}
		})
	}
}

// --- computeStaffHours tests ---

func TestComputeStaffHours(t *testing.T) {
	tests := []struct {
		name    string
		entries []model.LLMShiftEntry
		want    map[string]float64
	}{
		{
			name: "single entry no break",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "09:00", EndTime: "17:00", BreakMinutes: 0},
			},
			want: map[string]float64{"s1": 8.0},
		},
		{
			name: "single entry with 60min break",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "09:00", EndTime: "17:00", BreakMinutes: 60},
			},
			want: map[string]float64{"s1": 7.0},
		},
		{
			name: "multiple entries same staff",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "09:00", EndTime: "17:00", BreakMinutes: 60},
				{StaffID: "s1", Date: "2025-01-02", StartTime: "09:00", EndTime: "17:00", BreakMinutes: 60},
			},
			want: map[string]float64{"s1": 14.0},
		},
		{
			name: "multiple staff",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "09:00", EndTime: "17:00", BreakMinutes: 60},
				{StaffID: "s2", Date: "2025-01-01", StartTime: "10:00", EndTime: "18:00", BreakMinutes: 60},
			},
			want: map[string]float64{"s1": 7.0, "s2": 7.0},
		},
		{
			name:    "empty entries",
			entries: []model.LLMShiftEntry{},
			want:    map[string]float64{},
		},
		{
			name: "break exceeds work time (clamp to 0)",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "09:00", EndTime: "10:00", BreakMinutes: 120},
			},
			want: map[string]float64{"s1": 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeStaffHours(tt.entries)
			if len(got) != len(tt.want) {
				t.Fatalf("computeStaffHours returned %d entries, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if g, ok := got[k]; !ok {
					t.Errorf("missing key %q", k)
				} else if g != v {
					t.Errorf("computeStaffHours[%q] = %f, want %f", k, g, v)
				}
			}
		})
	}
}

// --- checkConsecutiveDays tests ---

func TestCheckConsecutiveDays(t *testing.T) {
	v := &ShiftValidator{}

	tests := []struct {
		name           string
		entries        []model.LLMShiftEntry
		maxDays        int
		constraintType string
		wantViolations int
		wantIsValid    bool
	}{
		{
			name: "no consecutive days issue",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s1", Date: "2025-01-03"},
				{StaffID: "s1", Date: "2025-01-05"},
			},
			maxDays:        5,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
		{
			name: "exactly at limit (5 consecutive, max 5)",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s1", Date: "2025-01-02"},
				{StaffID: "s1", Date: "2025-01-03"},
				{StaffID: "s1", Date: "2025-01-04"},
				{StaffID: "s1", Date: "2025-01-05"},
			},
			maxDays:        5,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
		{
			name: "exceeds limit (6 consecutive, max 5)",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s1", Date: "2025-01-02"},
				{StaffID: "s1", Date: "2025-01-03"},
				{StaffID: "s1", Date: "2025-01-04"},
				{StaffID: "s1", Date: "2025-01-05"},
				{StaffID: "s1", Date: "2025-01-06"},
			},
			maxDays:        5,
			constraintType: "hard",
			wantViolations: 1,
			wantIsValid:    false,
		},
		{
			name: "soft constraint violation does not invalidate",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s1", Date: "2025-01-02"},
				{StaffID: "s1", Date: "2025-01-03"},
				{StaffID: "s1", Date: "2025-01-04"},
			},
			maxDays:        3,
			constraintType: "soft",
			wantViolations: 1,
			wantIsValid:    true,
		},
		{
			name: "multiple staff separate tracking",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s1", Date: "2025-01-02"},
				{StaffID: "s1", Date: "2025-01-03"},
				{StaffID: "s2", Date: "2025-01-01"},
				{StaffID: "s2", Date: "2025-01-02"},
			},
			maxDays:        5,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{"max_days": float64(tt.maxDays)}
			c := constraintData{
				Name:     "連続勤務制限",
				Type:     tt.constraintType,
				Category: "max_consecutive_days",
				Config:   mustMarshalJSON(config),
			}
			result := &model.ValidationResult{
				IsValid:    true,
				Violations: []model.Violation{},
			}

			v.checkConsecutiveDays(tt.entries, config, c, result)

			if len(result.Violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d", len(result.Violations), tt.wantViolations)
			}
			if result.IsValid != tt.wantIsValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantIsValid)
			}
		})
	}
}

// --- checkMinStaff tests ---

func TestCheckMinStaff(t *testing.T) {
	v := &ShiftValidator{}

	tests := []struct {
		name           string
		entries        []model.LLMShiftEntry
		minCount       int
		constraintType string
		wantViolations int
		wantIsValid    bool
	}{
		{
			name: "meets minimum",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s2", Date: "2025-01-01"},
			},
			minCount:       2,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
		{
			name: "below minimum",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
			},
			minCount:       2,
			constraintType: "hard",
			wantViolations: 1,
			wantIsValid:    false,
		},
		{
			name: "soft violation below minimum",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
			},
			minCount:       2,
			constraintType: "soft",
			wantViolations: 1,
			wantIsValid:    true,
		},
		{
			name: "multiple dates some below",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s2", Date: "2025-01-01"},
				{StaffID: "s1", Date: "2025-01-02"},
			},
			minCount:       2,
			constraintType: "hard",
			wantViolations: 1, // 01-02 has only 1 staff
			wantIsValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{"min_count": float64(tt.minCount)}
			c := constraintData{
				Name:     "最低人数",
				Type:     tt.constraintType,
				Category: "min_staff",
				Config:   mustMarshalJSON(config),
			}
			result := &model.ValidationResult{
				IsValid:    true,
				Violations: []model.Violation{},
			}

			v.checkMinStaff(tt.entries, config, c, result)

			if len(result.Violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d", len(result.Violations), tt.wantViolations)
			}
			if result.IsValid != tt.wantIsValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantIsValid)
			}
		})
	}
}

// --- checkMaxStaff tests ---

func TestCheckMaxStaff(t *testing.T) {
	v := &ShiftValidator{}

	tests := []struct {
		name           string
		entries        []model.LLMShiftEntry
		maxCount       int
		constraintType string
		wantViolations int
		wantIsValid    bool
	}{
		{
			name: "within limit",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s2", Date: "2025-01-01"},
			},
			maxCount:       3,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
		{
			name: "exceeds limit",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s2", Date: "2025-01-01"},
				{StaffID: "s3", Date: "2025-01-01"},
				{StaffID: "s4", Date: "2025-01-01"},
			},
			maxCount:       3,
			constraintType: "hard",
			wantViolations: 1,
			wantIsValid:    false,
		},
		{
			name: "at exactly limit",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01"},
				{StaffID: "s2", Date: "2025-01-01"},
				{StaffID: "s3", Date: "2025-01-01"},
			},
			maxCount:       3,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{"max_count": float64(tt.maxCount)}
			c := constraintData{
				Name:     "最大人数",
				Type:     tt.constraintType,
				Category: "max_staff",
				Config:   mustMarshalJSON(config),
			}
			result := &model.ValidationResult{
				IsValid:    true,
				Violations: []model.Violation{},
			}

			v.checkMaxStaff(tt.entries, config, c, result)

			if len(result.Violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d", len(result.Violations), tt.wantViolations)
			}
			if result.IsValid != tt.wantIsValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantIsValid)
			}
		})
	}
}

// --- checkRestHours tests ---

func TestCheckRestHours(t *testing.T) {
	v := &ShiftValidator{}

	tests := []struct {
		name           string
		entries        []model.LLMShiftEntry
		minHours       float64
		constraintType string
		wantViolations int
		wantIsValid    bool
	}{
		{
			name: "sufficient rest interval (13h)",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "09:00", EndTime: "17:00"},
				{StaffID: "s1", Date: "2025-01-02", StartTime: "06:00", EndTime: "14:00"},
			},
			minHours:       11,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
		{
			name: "insufficient rest interval (8h)",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "14:00", EndTime: "22:00"},
				{StaffID: "s1", Date: "2025-01-02", StartTime: "06:00", EndTime: "14:00"},
			},
			minHours:       11,
			constraintType: "hard",
			wantViolations: 1,
			wantIsValid:    false,
		},
		{
			name: "exactly at limit (11h)",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "09:00", EndTime: "20:00"},
				{StaffID: "s1", Date: "2025-01-02", StartTime: "07:00", EndTime: "15:00"},
			},
			minHours:       11,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
		{
			name: "non-consecutive dates - no check",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "14:00", EndTime: "23:00"},
				{StaffID: "s1", Date: "2025-01-03", StartTime: "06:00", EndTime: "14:00"},
			},
			minHours:       11,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
		{
			name: "soft constraint violation",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "14:00", EndTime: "22:00"},
				{StaffID: "s1", Date: "2025-01-02", StartTime: "06:00", EndTime: "14:00"},
			},
			minHours:       11,
			constraintType: "soft",
			wantViolations: 1,
			wantIsValid:    true,
		},
		{
			name: "different staff - independent",
			entries: []model.LLMShiftEntry{
				{StaffID: "s1", Date: "2025-01-01", StartTime: "14:00", EndTime: "22:00"},
				{StaffID: "s2", Date: "2025-01-02", StartTime: "06:00", EndTime: "14:00"},
			},
			minHours:       11,
			constraintType: "hard",
			wantViolations: 0,
			wantIsValid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]interface{}{"min_hours": tt.minHours}
			c := constraintData{
				Name:     "勤務間インターバル",
				Type:     tt.constraintType,
				Category: "rest_hours",
				Config:   mustMarshalJSON(config),
			}
			result := &model.ValidationResult{
				IsValid:    true,
				Violations: []model.Violation{},
			}

			v.checkRestHours(tt.entries, config, c, result)

			if len(result.Violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d", len(result.Violations), tt.wantViolations)
			}
			if result.IsValid != tt.wantIsValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantIsValid)
			}
		})
	}
}

func mustMarshalJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
