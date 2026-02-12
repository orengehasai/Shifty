package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
	"shift-app/internal/service"
)

type ConstraintHandler struct {
	svc *service.ConstraintService
}

func NewConstraintHandler(svc *service.ConstraintService) *ConstraintHandler {
	return &ConstraintHandler{svc: svc}
}

func (h *ConstraintHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/constraints", h.List)
	g.POST("/constraints", h.Create)
	g.PUT("/constraints/:id", h.Update)
	g.DELETE("/constraints/:id", h.Delete)
}

func (h *ConstraintHandler) List(c echo.Context) error {
	isActive := parseBoolParam(c.QueryParam("is_active"))
	cType := parseStringParam(c.QueryParam("type"))
	category := parseStringParam(c.QueryParam("category"))

	constraints, err := h.svc.List(c.Request().Context(), isActive, cType, category)
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"constraints": constraints,
	})
}

func (h *ConstraintHandler) Create(c echo.Context) error {
	var req model.CreateConstraintRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	constraint, err := h.svc.Create(c.Request().Context(), req)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	return c.JSON(http.StatusCreated, constraint)
}

func (h *ConstraintHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req model.UpdateConstraintRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	constraint, err := h.svc.Update(c.Request().Context(), id, req)
	if err != nil {
		return internalError(c, err)
	}
	if constraint == nil {
		return notFound(c, "制約条件")
	}
	return c.JSON(http.StatusOK, constraint)
}

func (h *ConstraintHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return internalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
