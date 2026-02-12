package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
	"shift-app/internal/service"
)

type StaffHandler struct {
	svc *service.StaffService
}

func NewStaffHandler(svc *service.StaffService) *StaffHandler {
	return &StaffHandler{svc: svc}
}

func (h *StaffHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/staffs", h.List)
	g.POST("/staffs", h.Create)
	g.GET("/staffs/:id", h.GetByID)
	g.PUT("/staffs/:id", h.Update)
	g.DELETE("/staffs/:id", h.Delete)
}

func (h *StaffHandler) List(c echo.Context) error {
	isActive := parseBoolParam(c.QueryParam("is_active"))
	staffs, err := h.svc.List(c.Request().Context(), isActive)
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"staffs": staffs,
	})
}

func (h *StaffHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	staff, err := h.svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return internalError(c, err)
	}
	if staff == nil {
		return notFound(c, "スタッフ")
	}
	return c.JSON(http.StatusOK, staff)
}

func (h *StaffHandler) Create(c echo.Context) error {
	var req model.CreateStaffRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	staff, err := h.svc.Create(c.Request().Context(), req)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	return c.JSON(http.StatusCreated, staff)
}

func (h *StaffHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req model.UpdateStaffRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	staff, err := h.svc.Update(c.Request().Context(), id, req)
	if err != nil {
		return internalError(c, err)
	}
	if staff == nil {
		return notFound(c, "スタッフ")
	}
	return c.JSON(http.StatusOK, staff)
}

func (h *StaffHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return internalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
