package model

import (
	"encoding/json"
	"time"
)

// Staff represents the staffs table
type Staff struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	EmploymentType string    `json:"employment_type"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// StaffMonthlySetting represents the staff_monthly_settings table
type StaffMonthlySetting struct {
	ID                string    `json:"id"`
	StaffID           string    `json:"staff_id"`
	StaffName         string    `json:"staff_name,omitempty"`
	YearMonth         string    `json:"year_month"`
	MinPreferredHours int       `json:"min_preferred_hours"`
	MaxPreferredHours int       `json:"max_preferred_hours"`
	Note              *string   `json:"note"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ShiftRequest represents the shift_requests table
type ShiftRequest struct {
	ID          string    `json:"id"`
	StaffID     string    `json:"staff_id"`
	StaffName   string    `json:"staff_name,omitempty"`
	YearMonth   string    `json:"year_month"`
	Date        string    `json:"date"`
	StartTime   *string   `json:"start_time"`
	EndTime     *string   `json:"end_time"`
	RequestType string    `json:"request_type"`
	Note        *string   `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Constraint represents the constraints table
type Constraint struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Category  string          `json:"category"`
	Config    json.RawMessage `json:"config"`
	IsActive  bool            `json:"is_active"`
	Priority  *int            `json:"priority"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ConstraintViolation represents a single constraint violation
type ConstraintViolation struct {
	ConstraintName string `json:"constraint_name"`
	Type           string `json:"type"`
	Message        string `json:"message"`
}

// ShiftPattern represents the shift_patterns table
type ShiftPattern struct {
	ID                   string                `json:"id"`
	YearMonth            string                `json:"year_month"`
	Status               string                `json:"status"`
	Reasoning            *string               `json:"reasoning"`
	Score                *float64              `json:"score"`
	ConstraintViolations []ConstraintViolation `json:"constraint_violations"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at,omitempty"`
}

// ShiftEntry represents the shift_entries table
type ShiftEntry struct {
	ID           string    `json:"id"`
	PatternID    string    `json:"pattern_id"`
	StaffID      string    `json:"staff_id"`
	StaffName    string    `json:"staff_name,omitempty"`
	Date         string    `json:"date"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	BreakMinutes int       `json:"break_minutes"`
	IsManualEdit bool      `json:"is_manual_edit"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GenerationJob represents the generation_jobs table
type GenerationJob struct {
	ID            string     `json:"id"`
	YearMonth     string     `json:"year_month"`
	Status        string     `json:"status"`
	PatternCount  int        `json:"pattern_count"`
	Progress      int        `json:"progress"`
	StatusMessage *string    `json:"status_message"`
	ErrorMessage  *string    `json:"error_message"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// --- Request / Response DTOs ---

// CreateStaffRequest is the request body for POST /staffs
type CreateStaffRequest struct {
	Name           string `json:"name"`
	Role           string `json:"role"`
	EmploymentType string `json:"employment_type"`
}

// UpdateStaffRequest is the request body for PUT /staffs/:id
type UpdateStaffRequest struct {
	Name           *string `json:"name"`
	Role           *string `json:"role"`
	EmploymentType *string `json:"employment_type"`
	IsActive       *bool   `json:"is_active"`
}

// CreateStaffMonthlySettingRequest is the request body for POST /staff-monthly-settings
type CreateStaffMonthlySettingRequest struct {
	StaffID           string  `json:"staff_id"`
	YearMonth         string  `json:"year_month"`
	MinPreferredHours int     `json:"min_preferred_hours"`
	MaxPreferredHours int     `json:"max_preferred_hours"`
	Note              *string `json:"note"`
}

// BatchStaffMonthlySettingRequest is the request body for POST /staff-monthly-settings/batch
type BatchStaffMonthlySettingRequest struct {
	Settings []CreateStaffMonthlySettingRequest `json:"settings"`
}

// CreateShiftRequestRequest is the request body for POST /shift-requests
type CreateShiftRequestRequest struct {
	StaffID     string  `json:"staff_id"`
	YearMonth   string  `json:"year_month"`
	Date        string  `json:"date"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	RequestType string  `json:"request_type"`
	Note        *string `json:"note"`
}

// BatchShiftRequestRequest is the request body for POST /shift-requests/batch
type BatchShiftRequestRequest struct {
	Requests []CreateShiftRequestRequest `json:"requests"`
}

// CreateConstraintRequest is the request body for POST /constraints
type CreateConstraintRequest struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Category string          `json:"category"`
	Config   json.RawMessage `json:"config"`
	Priority *int            `json:"priority"`
}

// UpdateConstraintRequest is the request body for PUT /constraints/:id
type UpdateConstraintRequest struct {
	Name     *string          `json:"name"`
	Type     *string          `json:"type"`
	Category *string          `json:"category"`
	Config   *json.RawMessage `json:"config"`
	IsActive *bool            `json:"is_active"`
	Priority *int             `json:"priority"`
}

// GenerateShiftRequest is the request body for POST /shifts/generate
type GenerateShiftRequest struct {
	YearMonth    string `json:"year_month"`
	PatternCount int    `json:"pattern_count"`
}

// CreateShiftEntryRequest is the request body for POST /shifts/entries
type CreateShiftEntryRequest struct {
	PatternID    string `json:"pattern_id"`
	StaffID      string `json:"staff_id"`
	Date         string `json:"date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	BreakMinutes int    `json:"break_minutes"`
}

// UpdateShiftEntryRequest is the request body for PUT /shifts/entries/:id
type UpdateShiftEntryRequest struct {
	StartTime    *string `json:"start_time"`
	EndTime      *string `json:"end_time"`
	BreakMinutes *int    `json:"break_minutes"`
}

// --- API Response structures ---

// ErrorDetail represents a field-level validation error
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse represents the standard error response
type ErrorResponse struct {
	Error struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Details []ErrorDetail `json:"details,omitempty"`
	} `json:"error"`
}

// DashboardSummary represents GET /dashboard/summary response
type DashboardSummary struct {
	YearMonth             string           `json:"year_month"`
	StaffCount            int              `json:"staff_count"`
	ActiveStaffCount      int              `json:"active_staff_count"`
	RequestSubmittedCount int              `json:"request_submitted_count"`
	MonthlySettingsCount  int              `json:"monthly_settings_count"`
	ShiftStatus           string           `json:"shift_status"`
	ConstraintCount       int              `json:"constraint_count"`
	DailyStaffCounts      []DailyStaffCount `json:"daily_staff_counts"`
}

// DailyStaffCount represents a date with its staff count
type DailyStaffCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// PatternWithEntries is a pattern including its shift entries
type PatternWithEntries struct {
	ShiftPattern
	Entries []ShiftEntry `json:"entries"`
}

// PatternWithSummary is a pattern with computed summary data
type PatternWithSummary struct {
	ShiftPattern
	Summary *PatternSummary `json:"summary"`
}

// PatternSummary is the computed summary for a pattern
type PatternSummary struct {
	TotalEntries int                `json:"total_entries"`
	StaffHours   map[string]float64 `json:"staff_hours"`
}

// EntryValidation is returned alongside an entry update
type EntryValidation struct {
	IsValid  bool              `json:"is_valid"`
	Warnings []ValidationWarning `json:"warnings"`
}

// ValidationWarning represents a soft constraint warning
type ValidationWarning struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// --- LLM types ---

// LLMResponse represents the structured JSON output from Claude
type LLMResponse struct {
	Reasoning            string                `json:"reasoning"`
	Entries              []LLMShiftEntry       `json:"entries"`
	ConstraintViolations []ConstraintViolation `json:"constraint_violations"`
}

// LLMShiftEntry represents a single shift entry from LLM output
type LLMShiftEntry struct {
	StaffID      string `json:"staff_id"`
	Date         string `json:"date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	BreakMinutes int    `json:"break_minutes"`
}

// ValidationResult is returned by the shift validator
type ValidationResult struct {
	IsValid    bool        `json:"is_valid"`
	Violations []Violation `json:"violations"`
	Warnings   []Warning   `json:"warnings"`
	Score      float64     `json:"score"`
}

// HasHardViolations returns true if there are any hard constraint violations
func (v *ValidationResult) HasHardViolations() bool {
	for _, viol := range v.Violations {
		if viol.Type == "hard" {
			return true
		}
	}
	return false
}

// Violation represents a constraint violation
type Violation struct {
	Type       string `json:"type"`
	Constraint string `json:"constraint"`
	Date       string `json:"date,omitempty"`
	StaffID    string `json:"staff_id,omitempty"`
	Message    string `json:"message"`
}

// Warning represents a soft constraint warning
type Warning struct {
	Type       string `json:"type"`
	Constraint string `json:"constraint"`
	Message    string `json:"message"`
}
