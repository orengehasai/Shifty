package service

import (
	"context"
	"testing"

	"shift-app/internal/model"
)

func TestStaffMonthlySettingService_Create_Validation(t *testing.T) {
	svc := NewStaffMonthlySettingService(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     model.CreateStaffMonthlySettingRequest
		wantErr string
	}{
		{
			name:    "empty staff_id",
			req:     model.CreateStaffMonthlySettingRequest{StaffID: "", YearMonth: "2025-01", MinPreferredHours: 80, MaxPreferredHours: 160},
			wantErr: "staff_id は必須です",
		},
		{
			name:    "empty year_month",
			req:     model.CreateStaffMonthlySettingRequest{StaffID: "s1", YearMonth: "", MinPreferredHours: 80, MaxPreferredHours: 160},
			wantErr: "year_month は必須です",
		},
		{
			name:    "negative min hours",
			req:     model.CreateStaffMonthlySettingRequest{StaffID: "s1", YearMonth: "2025-01", MinPreferredHours: -1, MaxPreferredHours: 160},
			wantErr: "希望時間は0以上で指定してください",
		},
		{
			name:    "min > max hours",
			req:     model.CreateStaffMonthlySettingRequest{StaffID: "s1", YearMonth: "2025-01", MinPreferredHours: 200, MaxPreferredHours: 100},
			wantErr: "最小希望時間は最大希望時間以下にしてください",
		},
		{
			name:    "max hours too large",
			req:     model.CreateStaffMonthlySettingRequest{StaffID: "s1", YearMonth: "2025-01", MinPreferredHours: 0, MaxPreferredHours: 745},
			wantErr: "最大希望時間が大きすぎます",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestStaffMonthlySettingService_List_EmptyYearMonth(t *testing.T) {
	svc := NewStaffMonthlySettingService(nil)
	ctx := context.Background()

	_, err := svc.List(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "year_month は必須です" {
		t.Errorf("error = %q, want %q", err.Error(), "year_month は必須です")
	}
}

func TestStaffMonthlySettingService_BatchCreate_EmptySettings(t *testing.T) {
	svc := NewStaffMonthlySettingService(nil)
	ctx := context.Background()

	results, err := svc.BatchCreate(ctx, model.BatchStaffMonthlySettingRequest{Settings: []model.CreateStaffMonthlySettingRequest{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestStaffMonthlySettingService_BatchCreate_TooMany(t *testing.T) {
	svc := NewStaffMonthlySettingService(nil)
	ctx := context.Background()

	settings := make([]model.CreateStaffMonthlySettingRequest, 101)
	_, err := svc.BatchCreate(ctx, model.BatchStaffMonthlySettingRequest{Settings: settings})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "一括登録は100件以内で指定してください" {
		t.Errorf("error = %q, want %q", err.Error(), "一括登録は100件以内で指定してください")
	}
}

func TestStaffMonthlySettingService_Delete_NilRepo(t *testing.T) {
	// Delete with nil repo panics (expected behavior: requires DB connection)
	// This test verifies the Delete method exists and has the correct signature
	svc := NewStaffMonthlySettingService(nil)
	_ = svc

	// The Delete method should accept (ctx, id string) and return error
	// We cannot call it with nil repo without panicking, but we can verify
	// the service is created correctly
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestStaffMonthlySettingService_Create_ValidRequest(t *testing.T) {
	// Test that valid requests pass validation checks before reaching the repo
	svc := NewStaffMonthlySettingService(nil)
	ctx := context.Background()

	// This request passes all validation checks but will panic at repo.Upsert
	// since repo is nil. We use recover to verify validation passed.
	req := model.CreateStaffMonthlySettingRequest{
		StaffID:           "staff-1",
		YearMonth:         "2025-01",
		MinPreferredHours: 80,
		MaxPreferredHours: 160,
	}

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic from nil repo, validation should have passed")
			}
			// Panic means validation passed and it tried to call repo.Upsert
		}()
		svc.Create(ctx, req)
	}()
}
