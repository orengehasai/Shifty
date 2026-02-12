package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shift-app/internal/service"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/dashboard/summary", h.Summary)
}

func (h *DashboardHandler) Summary(c echo.Context) error {
	yearMonth := c.QueryParam("year_month")
	if yearMonth == "" {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "year_month は必須です")
	}

	summary, err := h.svc.GetSummary(c.Request().Context(), yearMonth)
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(http.StatusOK, summary)
}
