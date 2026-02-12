package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
	"shift-app/internal/service"
)

type StaffMonthlySettingHandler struct {
	svc *service.StaffMonthlySettingService
}

func NewStaffMonthlySettingHandler(svc *service.StaffMonthlySettingService) *StaffMonthlySettingHandler {
	return &StaffMonthlySettingHandler{svc: svc}
}

func (h *StaffMonthlySettingHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/staff-monthly-settings", h.List)
	g.POST("/staff-monthly-settings", h.Create)
	g.POST("/staff-monthly-settings/batch", h.BatchCreate)
	g.PUT("/staff-monthly-settings/:id", h.Update)
	g.DELETE("/staff-monthly-settings/:id", h.Delete)
}

func (h *StaffMonthlySettingHandler) List(c echo.Context) error {
	yearMonth := c.QueryParam("year_month")
	if yearMonth == "" {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "year_month は必須です")
	}
	staffID := parseStringParam(c.QueryParam("staff_id"))

	settings, err := h.svc.List(c.Request().Context(), yearMonth, staffID)
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"settings": settings,
	})
}

func (h *StaffMonthlySettingHandler) Create(c echo.Context) error {
	var req model.CreateStaffMonthlySettingRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	setting, err := h.svc.Create(c.Request().Context(), req)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	return c.JSON(http.StatusCreated, setting)
}

func (h *StaffMonthlySettingHandler) BatchCreate(c echo.Context) error {
	var req model.BatchStaffMonthlySettingRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	settings, err := h.svc.BatchCreate(c.Request().Context(), req)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"created_count": len(settings),
		"settings":      settings,
	})
}

func (h *StaffMonthlySettingHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req model.CreateStaffMonthlySettingRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	setting, err := h.svc.Update(c.Request().Context(), id, req)
	if err != nil {
		return internalError(c, err)
	}
	if setting == nil {
		return notFound(c, "月間設定")
	}
	return c.JSON(http.StatusOK, setting)
}

func (h *StaffMonthlySettingHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return internalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
