package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
)

func TestParseBoolParam(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *bool
	}{
		{"empty string returns nil", "", nil},
		{"true returns true", "true", boolPtr(true)},
		{"false returns false", "false", boolPtr(false)},
		{"random string returns false", "random", boolPtr(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBoolParam(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseBoolParam(%q) = %v, want nil", tt.input, *got)
				}
			} else {
				if got == nil {
					t.Errorf("parseBoolParam(%q) = nil, want %v", tt.input, *tt.want)
				} else if *got != *tt.want {
					t.Errorf("parseBoolParam(%q) = %v, want %v", tt.input, *got, *tt.want)
				}
			}
		})
	}
}

func TestParseStringParam(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *string
	}{
		{"empty string returns nil", "", nil},
		{"non-empty string returns pointer", "hello", strPtr("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStringParam(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseStringParam(%q) = %v, want nil", tt.input, *got)
				}
			} else {
				if got == nil {
					t.Errorf("parseStringParam(%q) = nil, want %v", tt.input, *tt.want)
				} else if *got != *tt.want {
					t.Errorf("parseStringParam(%q) = %v, want %v", tt.input, *got, *tt.want)
				}
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := errorResponse(c, http.StatusBadRequest, "TEST_ERROR", "テストエラーメッセージ")
	if err != nil {
		t.Fatalf("errorResponse returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error.Code != "TEST_ERROR" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "TEST_ERROR")
	}
	if resp.Error.Message != "テストエラーメッセージ" {
		t.Errorf("error message = %q, want %q", resp.Error.Message, "テストエラーメッセージ")
	}
}

func TestValidationError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	details := []model.ErrorDetail{
		{Field: "name", Message: "名前は必須です"},
		{Field: "role", Message: "役割は必須です"},
	}

	err := validationError(c, details)
	if err != nil {
		t.Fatalf("validationError returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "VALIDATION_ERROR")
	}
	if resp.Error.Message != "名前は必須です" {
		t.Errorf("error message = %q, want %q", resp.Error.Message, "名前は必須です")
	}
	if len(resp.Error.Details) != 2 {
		t.Errorf("details count = %d, want %d", len(resp.Error.Details), 2)
	}
}

func TestNotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := notFound(c, "スタッフ")
	if err != nil {
		t.Fatalf("notFound returned error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error.Message != "スタッフが見つかりません" {
		t.Errorf("error message = %q, want %q", resp.Error.Message, "スタッフが見つかりません")
	}
}

func TestInternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := internalError(c, nil)
	if err != nil {
		t.Fatalf("internalError returned error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "INTERNAL_ERROR")
	}
}

func boolPtr(b bool) *bool     { return &b }
func strPtr(s string) *string  { return &s }
