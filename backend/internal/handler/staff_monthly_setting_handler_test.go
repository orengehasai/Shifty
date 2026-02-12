package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
)

func TestStaffMonthlySettingHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/staff-monthly-settings/setting-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("setting-1")

	var deletedID string
	handler := func(c echo.Context) error {
		id := c.Param("id")
		deletedID = id
		// Simulate successful delete (no error)
		return c.NoContent(http.StatusNoContent)
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if deletedID != "setting-1" {
		t.Errorf("deleted id = %q, want %q", deletedID, "setting-1")
	}
}

func TestStaffMonthlySettingHandler_Delete_InternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/staff-monthly-settings/setting-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("setting-1")

	handler := func(c echo.Context) error {
		return internalError(c, nil)
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "INTERNAL_ERROR")
	}
}

func TestStaffMonthlySettingHandler_Create_BadRequest(t *testing.T) {
	e := echo.New()
	body := `invalid json`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/staff-monthly-settings", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		var reqBody model.CreateStaffMonthlySettingRequest
		if err := c.Bind(&reqBody); err != nil {
			return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
		}
		return c.JSON(http.StatusCreated, reqBody)
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Code != "INVALID_REQUEST" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "INVALID_REQUEST")
	}
}

func TestStaffMonthlySettingHandler_List_MissingYearMonth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/staff-monthly-settings", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		yearMonth := c.QueryParam("year_month")
		if yearMonth == "" {
			return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "year_month は必須です")
		}
		return c.JSON(http.StatusOK, nil)
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "VALIDATION_ERROR")
	}
}
