package service

import (
	"context"

	"shift-app/internal/model"
	"shift-app/internal/repository"
)

type DashboardService struct {
	staffRepo      *repository.StaffRepository
	settingRepo    *repository.StaffMonthlySettingRepository
	requestRepo    *repository.ShiftRequestRepository
	constraintRepo *repository.ConstraintRepository
	patternRepo    *repository.ShiftPatternRepository
	entryRepo      *repository.ShiftEntryRepository
	jobRepo        *repository.GenerationJobRepository
}

func NewDashboardService(
	staffRepo *repository.StaffRepository,
	settingRepo *repository.StaffMonthlySettingRepository,
	requestRepo *repository.ShiftRequestRepository,
	constraintRepo *repository.ConstraintRepository,
	patternRepo *repository.ShiftPatternRepository,
	entryRepo *repository.ShiftEntryRepository,
	jobRepo *repository.GenerationJobRepository,
) *DashboardService {
	return &DashboardService{
		staffRepo:      staffRepo,
		settingRepo:    settingRepo,
		requestRepo:    requestRepo,
		constraintRepo: constraintRepo,
		patternRepo:    patternRepo,
		entryRepo:      entryRepo,
		jobRepo:        jobRepo,
	}
}

func (s *DashboardService) GetSummary(ctx context.Context, yearMonth string) (*model.DashboardSummary, error) {
	staffCount, err := s.staffRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	activeCount, err := s.staffRepo.CountActive(ctx)
	if err != nil {
		return nil, err
	}

	requestCount, err := s.requestRepo.CountDistinctStaffByYearMonth(ctx, yearMonth)
	if err != nil {
		return nil, err
	}

	settingsCount, err := s.settingRepo.CountByYearMonth(ctx, yearMonth)
	if err != nil {
		return nil, err
	}

	constraintCount, err := s.constraintRepo.CountActive(ctx)
	if err != nil {
		return nil, err
	}

	shiftStatus := s.determineShiftStatus(ctx, yearMonth, requestCount)

	// Get daily staff counts from the best pattern (finalized > selected > draft)
	var dailyCounts []model.DailyStaffCount
	patterns, _ := s.patternRepo.ListByYearMonth(ctx, yearMonth)
	if len(patterns) > 0 {
		// Find the best pattern
		bestPattern := patterns[0]
		for _, p := range patterns {
			if p.Status == "finalized" {
				bestPattern = p
				break
			}
			if p.Status == "selected" {
				bestPattern = p
			}
		}
		dailyCounts, _ = s.entryRepo.DailyStaffCounts(ctx, bestPattern.ID)
	}
	if dailyCounts == nil {
		dailyCounts = []model.DailyStaffCount{}
	}

	return &model.DashboardSummary{
		YearMonth:             yearMonth,
		StaffCount:            staffCount,
		ActiveStaffCount:      activeCount,
		RequestSubmittedCount: requestCount,
		MonthlySettingsCount:  settingsCount,
		ShiftStatus:           shiftStatus,
		ConstraintCount:       constraintCount,
		DailyStaffCounts:      dailyCounts,
	}, nil
}

func (s *DashboardService) determineShiftStatus(ctx context.Context, yearMonth string, requestCount int) string {
	// Check if there's an active generation job
	hasJob, _ := s.jobRepo.HasProcessingJob(ctx, yearMonth)
	if hasJob {
		return "generating"
	}

	// Check pattern status
	patternStatus, _ := s.patternRepo.GetLatestStatusByYearMonth(ctx, yearMonth)
	switch patternStatus {
	case "finalized":
		return "finalized"
	case "selected":
		return "selected"
	case "draft":
		return "generated"
	}

	if requestCount > 0 {
		return "requests_submitted"
	}

	return "not_started"
}
