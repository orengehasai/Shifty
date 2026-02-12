package service

import (
	"context"
	"testing"

	"shift-app/internal/model"
)

func TestShiftRequestService_Create_Validation(t *testing.T) {
	svc := NewShiftRequestService(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     model.CreateShiftRequestRequest
		wantErr string
	}{
		{
			name:    "empty staff_id",
			req:     model.CreateShiftRequestRequest{StaffID: "", YearMonth: "2025-01", Date: "2025-01-01", RequestType: "unavailable"},
			wantErr: "staff_id は必須です",
		},
		{
			name:    "empty year_month",
			req:     model.CreateShiftRequestRequest{StaffID: "s1", YearMonth: "", Date: "2025-01-01", RequestType: "unavailable"},
			wantErr: "year_month は必須です",
		},
		{
			name:    "invalid year_month format",
			req:     model.CreateShiftRequestRequest{StaffID: "s1", YearMonth: "2025-1", Date: "2025-01-01", RequestType: "unavailable"},
			wantErr: "year_month は YYYY-MM 形式で指定してください",
		},
		{
			name:    "empty date",
			req:     model.CreateShiftRequestRequest{StaffID: "s1", YearMonth: "2025-01", Date: "", RequestType: "unavailable"},
			wantErr: "date は必須です",
		},
		{
			name:    "empty request_type",
			req:     model.CreateShiftRequestRequest{StaffID: "s1", YearMonth: "2025-01", Date: "2025-01-01", RequestType: ""},
			wantErr: "request_type は必須です",
		},
		{
			name:    "invalid request_type",
			req:     model.CreateShiftRequestRequest{StaffID: "s1", YearMonth: "2025-01", Date: "2025-01-01", RequestType: "invalid"},
			wantErr: "request_type は available, unavailable, preferred のいずれかで指定してください",
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

func TestShiftRequestService_List_EmptyYearMonth(t *testing.T) {
	svc := NewShiftRequestService(nil)
	ctx := context.Background()

	_, err := svc.List(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "year_month は必須です" {
		t.Errorf("error = %q, want %q", err.Error(), "year_month は必須です")
	}
}

func TestShiftRequestService_BatchCreate_EmptyRequests(t *testing.T) {
	svc := NewShiftRequestService(nil)
	ctx := context.Background()

	results, err := svc.BatchCreate(ctx, model.BatchShiftRequestRequest{Requests: []model.CreateShiftRequestRequest{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestShiftRequestService_BatchCreate_TooManyRequests(t *testing.T) {
	svc := NewShiftRequestService(nil)
	ctx := context.Background()

	requests := make([]model.CreateShiftRequestRequest, 101)
	_, err := svc.BatchCreate(ctx, model.BatchShiftRequestRequest{Requests: requests})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "一括登録は100件以内で指定してください" {
		t.Errorf("error = %q, want %q", err.Error(), "一括登録は100件以内で指定してください")
	}
}
