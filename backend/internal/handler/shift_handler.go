package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
	"shift-app/internal/service"
)

type ShiftHandler struct {
	svc *service.ShiftService
}

func NewShiftHandler(svc *service.ShiftService) *ShiftHandler {
	return &ShiftHandler{svc: svc}
}

func (h *ShiftHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/shifts/generate", h.Generate)
	g.GET("/shifts/generate/:job_id", h.GetJobStatus)
	g.GET("/shifts/patterns", h.ListPatterns)
	g.GET("/shifts/patterns/:id", h.GetPatternDetail)
	g.PUT("/shifts/patterns/:id/select", h.SelectPattern)
	g.PUT("/shifts/patterns/:id/finalize", h.FinalizePattern)
	g.POST("/shifts/entries", h.CreateEntry)
	g.PUT("/shifts/entries/:id", h.UpdateEntry)
	g.DELETE("/shifts/entries/:id", h.DeleteEntry)
}

func (h *ShiftHandler) Generate(c echo.Context) error {
	var req model.GenerateShiftRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	job, err := h.svc.StartGeneration(c.Request().Context(), req)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"job_id":  job.ID,
		"status":  job.Status,
		"message": "シフト生成を開始しました",
	})
}

func (h *ShiftHandler) GetJobStatus(c echo.Context) error {
	jobID := c.Param("job_id")
	job, err := h.svc.GetJob(c.Request().Context(), jobID)
	if err != nil {
		return internalError(c, err)
	}
	if job == nil {
		return notFound(c, "ジョブ")
	}
	return c.JSON(http.StatusOK, job)
}

func (h *ShiftHandler) ListPatterns(c echo.Context) error {
	yearMonth := c.QueryParam("year_month")
	if yearMonth == "" {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "year_month は必須です")
	}

	patterns, err := h.svc.ListPatterns(c.Request().Context(), yearMonth)
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"patterns": patterns,
	})
}

func (h *ShiftHandler) GetPatternDetail(c echo.Context) error {
	id := c.Param("id")
	pattern, err := h.svc.GetPatternDetail(c.Request().Context(), id)
	if err != nil {
		return internalError(c, err)
	}
	if pattern == nil {
		return notFound(c, "パターン")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"pattern": pattern,
	})
}

func (h *ShiftHandler) SelectPattern(c echo.Context) error {
	id := c.Param("id")
	pattern, err := h.svc.SelectPattern(c.Request().Context(), id)
	if err != nil {
		return internalError(c, err)
	}
	if pattern == nil {
		return notFound(c, "パターン")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"pattern": map[string]interface{}{
			"id":     pattern.ID,
			"status": pattern.Status,
		},
	})
}

func (h *ShiftHandler) FinalizePattern(c echo.Context) error {
	id := c.Param("id")
	pattern, err := h.svc.FinalizePattern(c.Request().Context(), id)
	if err != nil {
		return internalError(c, err)
	}
	if pattern == nil {
		return notFound(c, "パターン")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"pattern": map[string]interface{}{
			"id":     pattern.ID,
			"status": pattern.Status,
		},
	})
}

func (h *ShiftHandler) CreateEntry(c echo.Context) error {
	var req model.CreateShiftEntryRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	entry, err := h.svc.CreateEntry(c.Request().Context(), req)
	if err != nil {
		return internalError(c, err)
	}
	return c.JSON(http.StatusCreated, entry)
}

func (h *ShiftHandler) UpdateEntry(c echo.Context) error {
	id := c.Param("id")
	var req model.UpdateShiftEntryRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの形式が不正です")
	}

	entry, validation, err := h.svc.UpdateEntry(c.Request().Context(), id, req)
	if err != nil {
		return internalError(c, err)
	}
	if entry == nil {
		return notFound(c, "エントリ")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"entry":      entry,
		"validation": validation,
	})
}

func (h *ShiftHandler) DeleteEntry(c echo.Context) error {
	id := c.Param("id")
	if err := h.svc.DeleteEntry(c.Request().Context(), id); err != nil {
		return internalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
