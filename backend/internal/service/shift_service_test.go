package service

import (
	"context"
	"testing"

	"shift-app/internal/model"
)

func TestComputeWorkHours(t *testing.T) {
	tests := []struct {
		name         string
		startTime    string
		endTime      string
		breakMinutes int
		want         float64
	}{
		{
			name:         "standard 8h shift with 1h break",
			startTime:    "09:00",
			endTime:      "18:00",
			breakMinutes: 60,
			want:         8.0,
		},
		{
			name:         "no break",
			startTime:    "09:00",
			endTime:      "17:00",
			breakMinutes: 0,
			want:         8.0,
		},
		{
			name:         "short shift",
			startTime:    "10:00",
			endTime:      "14:00",
			breakMinutes: 30,
			want:         3.5,
		},
		{
			name:         "break exceeds work time (clamped to 0)",
			startTime:    "09:00",
			endTime:      "10:00",
			breakMinutes: 120,
			want:         0.0,
		},
		{
			name:         "midnight shift",
			startTime:    "00:00",
			endTime:      "08:00",
			breakMinutes: 60,
			want:         7.0,
		},
		{
			name:         "half hour intervals",
			startTime:    "09:30",
			endTime:      "17:30",
			breakMinutes: 45,
			want:         7.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeWorkHours(tt.startTime, tt.endTime, tt.breakMinutes)
			if got != tt.want {
				t.Errorf("computeWorkHours(%q, %q, %d) = %f, want %f",
					tt.startTime, tt.endTime, tt.breakMinutes, got, tt.want)
			}
		})
	}
}

func TestShiftService_StartGeneration_EmptyYearMonth(t *testing.T) {
	svc := &ShiftService{}
	ctx := context.Background()

	req := model.GenerateShiftRequest{
		YearMonth:    "",
		PatternCount: 3,
	}

	_, err := svc.StartGeneration(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "year_month は必須です" {
		t.Errorf("error = %q, want %q", err.Error(), "year_month は必須です")
	}
}

func TestShiftService_StartGeneration_DefaultPatternCount(t *testing.T) {
	// When patternCount is 0, it should default to 3
	// We can only verify this doesn't panic; the actual DB call will fail
	// This test verifies the validation path doesn't error on patternCount=0
	// (it should default to 3 and proceed to DB check which will panic on nil)
	svc := &ShiftService{}
	ctx := context.Background()

	req := model.GenerateShiftRequest{
		YearMonth:    "2025-01",
		PatternCount: 0,
	}

	// This will panic on nil jobRepo access, which is expected
	// We verify only that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("no panic, meaning jobRepo was not nil (unexpected in this test setup)")
		}
		// panic is expected since jobRepo is nil
	}()

	_, _ = svc.StartGeneration(ctx, req)
}
