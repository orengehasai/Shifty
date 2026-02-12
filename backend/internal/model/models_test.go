package model

import (
	"testing"
)

func TestValidationResult_HasHardViolations(t *testing.T) {
	tests := []struct {
		name       string
		violations []Violation
		want       bool
	}{
		{
			name:       "no violations",
			violations: []Violation{},
			want:       false,
		},
		{
			name: "only soft violations",
			violations: []Violation{
				{Type: "soft", Constraint: "test", Message: "msg"},
			},
			want: false,
		},
		{
			name: "has hard violation",
			violations: []Violation{
				{Type: "hard", Constraint: "test", Message: "msg"},
			},
			want: true,
		},
		{
			name: "mixed violations",
			violations: []Violation{
				{Type: "soft", Constraint: "test1", Message: "msg1"},
				{Type: "hard", Constraint: "test2", Message: "msg2"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &ValidationResult{
				Violations: tt.violations,
			}
			got := vr.HasHardViolations()
			if got != tt.want {
				t.Errorf("HasHardViolations() = %v, want %v", got, tt.want)
			}
		})
	}
}
