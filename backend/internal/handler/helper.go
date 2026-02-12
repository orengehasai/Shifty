package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"shift-app/internal/model"
)

func errorResponse(c echo.Context, status int, code string, message string) error {
	resp := model.ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	return c.JSON(status, resp)
}

func validationError(c echo.Context, details []model.ErrorDetail) error {
	resp := model.ErrorResponse{}
	resp.Error.Code = "VALIDATION_ERROR"
	if len(details) > 0 {
		resp.Error.Message = details[0].Message
	}
	resp.Error.Details = details
	return c.JSON(http.StatusBadRequest, resp)
}

func notFound(c echo.Context, resource string) error {
	return errorResponse(c, http.StatusNotFound, "NOT_FOUND", resource+"が見つかりません")
}

func internalError(c echo.Context, err error) error {
	log.Printf("[ERROR] %s %s: %v", c.Request().Method, c.Request().URL.Path, err)
	return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "内部エラーが発生しました")
}

func parseBoolParam(value string) *bool {
	if value == "" {
		return nil
	}
	b := value == "true"
	return &b
}

func parseStringParam(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
