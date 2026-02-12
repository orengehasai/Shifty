package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"shift-app/internal/config"
	"shift-app/internal/handler"
	"shift-app/internal/llm"
	"shift-app/internal/middleware"
	"shift-app/internal/repository"
	"shift-app/internal/service"
	"shift-app/internal/validator"
)

func main() {
	cfg := config.Load()

	// DB connection
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Repositories
	staffRepo := repository.NewStaffRepository(pool)
	settingRepo := repository.NewStaffMonthlySettingRepository(pool)
	requestRepo := repository.NewShiftRequestRepository(pool)
	constraintRepo := repository.NewConstraintRepository(pool)
	patternRepo := repository.NewShiftPatternRepository(pool)
	entryRepo := repository.NewShiftEntryRepository(pool)
	jobRepo := repository.NewGenerationJobRepository(pool)

	// LLM & Validator
	gen := llm.NewGenerator(cfg.AnthropicAPIKey, pool)
	val := validator.NewShiftValidator(pool)

	// Services
	staffSvc := service.NewStaffService(staffRepo)
	settingSvc := service.NewStaffMonthlySettingService(settingRepo)
	requestSvc := service.NewShiftRequestService(requestRepo)
	constraintSvc := service.NewConstraintService(constraintRepo)
	dashboardSvc := service.NewDashboardService(staffRepo, settingRepo, requestRepo, constraintRepo, patternRepo, entryRepo, jobRepo)
	shiftSvc := service.NewShiftService(patternRepo, entryRepo, jobRepo, staffRepo, gen, val)

	// Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.CORSConfig())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// API routes
	api := e.Group("/api/v1")

	// Register handlers
	staffHandler := handler.NewStaffHandler(staffSvc)
	staffHandler.RegisterRoutes(api)

	settingHandler := handler.NewStaffMonthlySettingHandler(settingSvc)
	settingHandler.RegisterRoutes(api)

	requestHandler := handler.NewShiftRequestHandler(requestSvc)
	requestHandler.RegisterRoutes(api)

	constraintHandler := handler.NewConstraintHandler(constraintSvc)
	constraintHandler.RegisterRoutes(api)

	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)
	dashboardHandler.RegisterRoutes(api)

	shiftHandler := handler.NewShiftHandler(shiftSvc)
	shiftHandler.RegisterRoutes(api)

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Starting server on %s", addr)
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
