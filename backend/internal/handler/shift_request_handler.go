package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
	"shift-app/internal/service"
)

type ShiftRequestHandler struct {
	svc *service.ShiftRequestService
}

func NewShiftRequestHandler(svc *service.ShiftRequestService) *ShiftRequestHandler {
	return &ShiftRequestHandler{svc: svc}
}

func (h *ShiftRequestHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/shift-requests", h.List)
	g.POST("/shift-requests", h.Create)
	g.POST("/shift-requests/batch", h.BatchCreate)
	g.PUT("/shift-requests/:id", h.Update)
	g.DELETE("/shift-requests/:id", h.Delete)
}

func (h *ShiftRequestHandler) List(c echo.Context) error {
	yearMonth := c.QueryParam("year_month")
	if yearMonth == "" {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "year_month は必須です")
	}
	staffID := parseStringParam(c.QueryParam("staff_id"))

	requests, err := h.svc.List(c.Request().Context(), yearMonth, staffID)
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"shift_requests": requests,
	})
}

func (h *ShiftRequestHandler) Create(c echo.Context) error {
	var req model.CreateShiftRequestRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	result, err := h.svc.Create(c.Request().Context(), req)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *ShiftRequestHandler) BatchCreate(c echo.Context) error {
	var req model.BatchShiftRequestRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	results, err := h.svc.BatchCreate(c.Request().Context(), req)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"created_count":  len(results),
		"shift_requests": results,
	})
}

func (h *ShiftRequestHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req model.CreateShiftRequestRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	result, err := h.svc.Update(c.Request().Context(), id, req)
	if err != nil {
		return internalError(c, err)
	}
	if result == nil {
		return notFound(c, "シフト希望")
	}
	return c.JSON(http.StatusOK, result)
}

func (h *ShiftRequestHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return internalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
