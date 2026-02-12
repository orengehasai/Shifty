package middleware

import (
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func CORSConfig() echo.MiddlewareFunc {
	origins := []string{"http://localhost:5173", "http://localhost:3000"}
	if env := os.Getenv("CORS_ALLOWED_ORIGINS"); env != "" {
		origins = strings.Split(env, ",")
	}
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	})
}
