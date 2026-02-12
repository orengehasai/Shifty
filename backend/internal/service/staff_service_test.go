package service

import (
	"context"
	"testing"

	"shift-app/internal/model"
)

func TestStaffService_Create_Validation(t *testing.T) {
	svc := NewStaffService(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     model.CreateStaffRequest
		wantErr string
	}{
		{
			name:    "empty name",
			req:     model.CreateStaffRequest{Name: "", Role: "kitchen", EmploymentType: "full_time"},
			wantErr: "名前は必須です",
		},
		{
			name:    "name too long",
			req:     model.CreateStaffRequest{Name: string(make([]byte, 101)), Role: "kitchen", EmploymentType: "full_time"},
			wantErr: "名前は100文字以内で入力してください",
		},
		{
			name:    "empty role",
			req:     model.CreateStaffRequest{Name: "田中", Role: "", EmploymentType: "full_time"},
			wantErr: "役割は必須です",
		},
		{
			name:    "invalid role",
			req:     model.CreateStaffRequest{Name: "田中", Role: "invalid_role", EmploymentType: "full_time"},
			wantErr: "役割は kitchen, hall, both のいずれかで指定してください",
		},
		{
			name:    "empty employment_type",
			req:     model.CreateStaffRequest{Name: "田中", Role: "kitchen", EmploymentType: ""},
			wantErr: "雇用形態は必須です",
		},
		{
			name:    "invalid employment_type",
			req:     model.CreateStaffRequest{Name: "田中", Role: "kitchen", EmploymentType: "contract"},
			wantErr: "雇用形態は full_time, part_time のいずれかで指定してください",
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

func TestStaffService_UpdateStaffRequest_IsActiveField(t *testing.T) {
	// UpdateStaffRequest includes IsActive *bool field
	// This test validates that the DTO correctly holds the is_active value
	// for staff reactivation (setting is_active=true on a soft-deleted staff)

	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		req      model.UpdateStaffRequest
		hasActive bool
		wantActive bool
	}{
		{
			name:      "is_active true for reactivation",
			req:       model.UpdateStaffRequest{IsActive: &trueVal},
			hasActive: true,
			wantActive: true,
		},
		{
			name:      "is_active false for soft-delete",
			req:       model.UpdateStaffRequest{IsActive: &falseVal},
			hasActive: true,
			wantActive: false,
		},
		{
			name:      "is_active nil means no change",
			req:       model.UpdateStaffRequest{IsActive: nil},
			hasActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.hasActive {
				if tt.req.IsActive == nil {
					t.Fatal("expected IsActive to be set, got nil")
				}
				if *tt.req.IsActive != tt.wantActive {
					t.Errorf("IsActive = %v, want %v", *tt.req.IsActive, tt.wantActive)
				}
			} else {
				if tt.req.IsActive != nil {
					t.Errorf("IsActive = %v, want nil", *tt.req.IsActive)
				}
			}
		})
	}
}
