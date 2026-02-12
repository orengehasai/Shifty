package service

import (
	"context"
	"testing"

	"shift-app/internal/model"
)

func TestConstraintService_Create_Validation(t *testing.T) {
	svc := NewConstraintService(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     model.CreateConstraintRequest
		wantErr string
	}{
		{
			name:    "empty name",
			req:     model.CreateConstraintRequest{Name: "", Type: "hard", Category: "min_staff"},
			wantErr: "名前は必須です",
		},
		{
			name:    "name too long",
			req:     model.CreateConstraintRequest{Name: string(make([]byte, 201)), Type: "hard", Category: "min_staff"},
			wantErr: "名前は200文字以内で入力してください",
		},
		{
			name:    "empty type",
			req:     model.CreateConstraintRequest{Name: "制約1", Type: "", Category: "min_staff"},
			wantErr: "type は必須です",
		},
		{
			name:    "invalid type",
			req:     model.CreateConstraintRequest{Name: "制約1", Type: "invalid", Category: "min_staff"},
			wantErr: "type は hard, soft のいずれかで指定してください",
		},
		{
			name:    "empty category",
			req:     model.CreateConstraintRequest{Name: "制約1", Type: "hard", Category: ""},
			wantErr: "category は必須です",
		},
		{
			name:    "invalid category",
			req:     model.CreateConstraintRequest{Name: "制約1", Type: "hard", Category: "invalid_cat"},
			wantErr: "無効な category です",
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
