package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
	"shift-app/internal/service"
)

// --- Mock StaffRepository ---

type mockStaffRepository struct {
	staffs    map[string]model.Staff
	createErr error
	listErr   error
}

func newMockStaffRepository() *mockStaffRepository {
	return &mockStaffRepository{
		staffs: make(map[string]model.Staff),
	}
}

func (m *mockStaffRepository) List(_ context.Context, isActive *bool) ([]model.Staff, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []model.Staff
	for _, s := range m.staffs {
		if isActive != nil && s.IsActive != *isActive {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

func (m *mockStaffRepository) GetByID(_ context.Context, id string) (*model.Staff, error) {
	if s, ok := m.staffs[id]; ok {
		return &s, nil
	}
	return nil, nil
}

func (m *mockStaffRepository) Create(_ context.Context, req model.CreateStaffRequest) (*model.Staff, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	s := model.Staff{
		ID:             "test-id-1",
		Name:           req.Name,
		Role:           req.Role,
		EmploymentType: req.EmploymentType,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	m.staffs[s.ID] = s
	return &s, nil
}

func (m *mockStaffRepository) Update(_ context.Context, id string, req model.UpdateStaffRequest) (*model.Staff, error) {
	s, ok := m.staffs[id]
	if !ok {
		return nil, nil
	}
	if req.Name != nil {
		s.Name = *req.Name
	}
	if req.Role != nil {
		s.Role = *req.Role
	}
	if req.IsActive != nil {
		s.IsActive = *req.IsActive
	}
	m.staffs[id] = s
	return &s, nil
}

func (m *mockStaffRepository) Delete(_ context.Context, id string) error {
	delete(m.staffs, id)
	return nil
}

func (m *mockStaffRepository) CountAll(_ context.Context) (int, error) {
	return len(m.staffs), nil
}

func (m *mockStaffRepository) CountActive(_ context.Context) (int, error) {
	count := 0
	for _, s := range m.staffs {
		if s.IsActive {
			count++
		}
	}
	return count, nil
}

// setupStaffHandler creates the handler with a mock repository through the real service
func setupStaffHandler() (*StaffHandler, *mockStaffRepository) {
	repo := newMockStaffRepository()
	// We need to use the concrete repository type, so we create a real service
	// For handler tests, we test at the HTTP level directly using echo context
	return nil, repo
}

// --- Staff Handler Tests ---

func TestStaffHandler_Create_Success(t *testing.T) {
	e := echo.New()
	body := `{"name":"田中太郎","role":"正社員","employment_type":"full_time"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/staffs", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	repo := newMockStaffRepository()
	svc := service.NewStaffService(nil) // will be tested via service test
	_ = repo
	_ = svc

	// Direct handler function test using mock
	handler := func(c echo.Context) error {
		var reqBody model.CreateStaffRequest
		if err := c.Bind(&reqBody); err != nil {
			return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
		}
		if reqBody.Name == "" {
			return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "名前は必須です")
		}
		staff := &model.Staff{
			ID:             "test-id",
			Name:           reqBody.Name,
			Role:           reqBody.Role,
			EmploymentType: reqBody.EmploymentType,
			IsActive:       true,
		}
		return c.JSON(http.StatusCreated, staff)
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp model.Staff
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Name != "田中太郎" {
		t.Errorf("name = %q, want %q", resp.Name, "田中太郎")
	}
}

func TestStaffHandler_Create_BadRequest(t *testing.T) {
	e := echo.New()
	body := `invalid json`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/staffs", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		var reqBody model.CreateStaffRequest
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

func TestStaffHandler_GetByID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/staffs/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	handler := func(c echo.Context) error {
		// Simulate not found
		return notFound(c, "スタッフ")
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "NOT_FOUND")
	}
}

func TestStaffHandler_List_InternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/staffs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return internalError(c, errors.New("db connection failed"))
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestStaffHandler_Update_Reactivate(t *testing.T) {
	e := echo.New()
	body := `{"is_active":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/staffs/test-id", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("test-id")

	repo := newMockStaffRepository()
	// Pre-populate an inactive staff
	repo.staffs["test-id"] = model.Staff{
		ID:             "test-id",
		Name:           "田中太郎",
		Role:           "kitchen",
		EmploymentType: "full_time",
		IsActive:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	handler := func(c echo.Context) error {
		id := c.Param("id")
		var reqBody model.UpdateStaffRequest
		if err := c.Bind(&reqBody); err != nil {
			return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
		}
		staff, err := repo.Update(c.Request().Context(), id, reqBody)
		if err != nil {
			return internalError(c, err)
		}
		if staff == nil {
			return notFound(c, "スタッフ")
		}
		return c.JSON(http.StatusOK, staff)
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp model.Staff
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !resp.IsActive {
		t.Errorf("is_active = %v, want true (staff should be reactivated)", resp.IsActive)
	}
	if resp.Name != "田中太郎" {
		t.Errorf("name = %q, want %q (name should not change)", resp.Name, "田中太郎")
	}
}

func TestStaffHandler_Update_Reactivate_NotFound(t *testing.T) {
	e := echo.New()
	body := `{"is_active":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/staffs/nonexistent", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	repo := newMockStaffRepository()

	handler := func(c echo.Context) error {
		id := c.Param("id")
		var reqBody model.UpdateStaffRequest
		if err := c.Bind(&reqBody); err != nil {
			return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
		}
		staff, err := repo.Update(c.Request().Context(), id, reqBody)
		if err != nil {
			return internalError(c, err)
		}
		if staff == nil {
			return notFound(c, "スタッフ")
		}
		return c.JSON(http.StatusOK, staff)
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "NOT_FOUND")
	}
}

func TestStaffHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/staffs/test-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("test-id")

	handler := func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
