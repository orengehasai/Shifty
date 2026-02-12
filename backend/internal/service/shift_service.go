package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"shift-app/internal/model"
	"shift-app/internal/repository"
)

// ShiftGenerator is an interface for the LLM generator
type ShiftGenerator interface {
	Generate(ctx context.Context, yearMonth string, patternCount int, patternIdx int, previousPatterns []model.LLMResponse, lastViolations []model.Violation) (*model.LLMResponse, error)
}

// ShiftValidator is an interface for the shift validator
type ShiftValidator interface {
	Validate(ctx context.Context, yearMonth string, response *model.LLMResponse) (*model.ValidationResult, error)
}

type ShiftService struct {
	patternRepo *repository.ShiftPatternRepository
	entryRepo   *repository.ShiftEntryRepository
	jobRepo     *repository.GenerationJobRepository
	staffRepo   *repository.StaffRepository
	generator   ShiftGenerator
	validator   ShiftValidator
}

func NewShiftService(
	patternRepo *repository.ShiftPatternRepository,
	entryRepo *repository.ShiftEntryRepository,
	jobRepo *repository.GenerationJobRepository,
	staffRepo *repository.StaffRepository,
	generator ShiftGenerator,
	validator ShiftValidator,
) *ShiftService {
	return &ShiftService{
		patternRepo: patternRepo,
		entryRepo:   entryRepo,
		jobRepo:     jobRepo,
		staffRepo:   staffRepo,
		generator:   generator,
		validator:   validator,
	}
}

func (s *ShiftService) StartGeneration(ctx context.Context, req model.GenerateShiftRequest) (*model.GenerationJob, error) {
	if req.YearMonth == "" {
		return nil, errors.New("year_month は必須です")
	}
	if len(req.YearMonth) != 7 {
		return nil, errors.New("year_month は YYYY-MM 形式で指定してください")
	}
	if req.PatternCount <= 0 {
		req.PatternCount = 3
	}
	if req.PatternCount > 5 {
		return nil, errors.New("パターン数は5以下で指定してください")
	}

	// Check for existing processing job
	hasJob, err := s.jobRepo.HasProcessingJob(ctx, req.YearMonth)
	if err != nil {
		return nil, err
	}
	if hasJob {
		return nil, errors.New("この月のシフト生成が既に進行中です")
	}

	job, err := s.jobRepo.Create(ctx, req.YearMonth, req.PatternCount)
	if err != nil {
		return nil, err
	}

	// Start async generation
	go s.runGeneration(job.ID, req.YearMonth, req.PatternCount)

	return job, nil
}

func (s *ShiftService) runGeneration(jobID string, yearMonth string, patternCount int) {
	ctx := context.Background()
	maxRetries := 3

	if err := s.jobRepo.SetProcessing(ctx, jobID); err != nil {
		log.Printf("Failed to set job processing: %v", err)
		return
	}

	var previousPatterns []model.LLMResponse

	for i := 0; i < patternCount; i++ {
		var finalResult *model.LLMResponse
		var finalValidation *model.ValidationResult
		var lastViolations []model.Violation

		for retry := 0; retry < maxRetries; retry++ {
			// Update progress: each pattern has equal weight, retries subdivide
			progressPct := (i * 100) / patternCount
			retryPct := (retry * 100) / (patternCount * maxRetries)
			progressPct += retryPct
			if progressPct > 95 {
				progressPct = 95
			}
			statusMsg := fmt.Sprintf("パターン%d/%d 生成中", i+1, patternCount)
			if retry > 0 {
				statusMsg = fmt.Sprintf("パターン%d/%d 再生成中 (試行%d/%d)", i+1, patternCount, retry+1, maxRetries)
			}
			_ = s.jobRepo.UpdateProgress(ctx, jobID, progressPct, statusMsg)

			result, err := s.generator.Generate(ctx, yearMonth, patternCount, i, previousPatterns, lastViolations)
			if err != nil {
				log.Printf("LLM generation failed (pattern %d, retry %d): %v", i+1, retry+1, err)
				if retry == maxRetries-1 {
					_ = s.jobRepo.SetFailed(ctx, jobID, fmt.Sprintf("パターン%d生成失敗: %v", i+1, err))
					return
				}
				continue
			}

			validation, err := s.validator.Validate(ctx, yearMonth, result)
			if err != nil {
				log.Printf("Validation failed: %v", err)
				if retry == maxRetries-1 {
					// Save even with validation errors
					finalResult = result
					finalValidation = validation
					break
				}
				continue
			}

			finalResult = result
			finalValidation = validation

			if !validation.HasHardViolations() {
				break
			}
			// Log violation details and pass them to next retry
			lastViolations = validation.Violations
			for _, v := range validation.Violations {
				if v.Type == "hard" {
					log.Printf("  [hard] %s: %s (staff=%s, date=%s)", v.Constraint, v.Message, v.StaffID, v.Date)
				}
			}
			log.Printf("Hard constraint violations found (pattern %d, retry %d), retrying with feedback...", i+1, retry+1)
		}

		if finalResult == nil {
			_ = s.jobRepo.SetFailed(ctx, jobID, fmt.Sprintf("パターン%d: 結果が空です", i+1))
			return
		}

		// Merge violations from LLM and validator
		violations := finalResult.ConstraintViolations
		if finalValidation != nil {
			for _, v := range finalValidation.Violations {
				violations = append(violations, model.ConstraintViolation{
					ConstraintName: v.Constraint,
					Type:           v.Type,
					Message:        v.Message,
				})
			}
		}

		score := float64(0)
		if finalValidation != nil {
			score = finalValidation.Score
		}

		pattern, err := s.patternRepo.Create(ctx, yearMonth, finalResult.Reasoning, score, violations)
		if err != nil {
			_ = s.jobRepo.SetFailed(ctx, jobID, fmt.Sprintf("パターン%d保存失敗: %v", i+1, err))
			return
		}

		if err := s.entryRepo.BulkCreate(ctx, pattern.ID, finalResult.Entries); err != nil {
			_ = s.jobRepo.SetFailed(ctx, jobID, fmt.Sprintf("エントリ保存失敗: %v", err))
			return
		}

		previousPatterns = append(previousPatterns, *finalResult)
	}

	_ = s.jobRepo.SetCompleted(ctx, jobID)
}

func (s *ShiftService) GetJob(ctx context.Context, jobID string) (*model.GenerationJob, error) {
	return s.jobRepo.GetByID(ctx, jobID)
}

func (s *ShiftService) ListPatterns(ctx context.Context, yearMonth string) ([]model.PatternWithSummary, error) {
	patterns, err := s.patternRepo.ListByYearMonth(ctx, yearMonth)
	if err != nil {
		return nil, err
	}

	var result []model.PatternWithSummary
	for _, p := range patterns {
		summary, _ := s.computeSummary(ctx, p.ID)
		result = append(result, model.PatternWithSummary{
			ShiftPattern: p,
			Summary:      summary,
		})
	}
	if result == nil {
		result = []model.PatternWithSummary{}
	}
	return result, nil
}

func (s *ShiftService) GetPatternDetail(ctx context.Context, id string) (*model.PatternWithEntries, error) {
	pattern, err := s.patternRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pattern == nil {
		return nil, nil
	}

	entries, err := s.entryRepo.ListByPatternID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []model.ShiftEntry{}
	}

	return &model.PatternWithEntries{
		ShiftPattern: *pattern,
		Entries:      entries,
	}, nil
}

func (s *ShiftService) SelectPattern(ctx context.Context, id string) (*model.ShiftPattern, error) {
	pattern, err := s.patternRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pattern == nil {
		return nil, nil
	}

	// Reset other patterns and set this one as selected
	if err := s.patternRepo.ResetOtherPatterns(ctx, id, pattern.YearMonth); err != nil {
		return nil, err
	}
	if err := s.patternRepo.UpdateStatus(ctx, id, "selected"); err != nil {
		return nil, err
	}

	pattern.Status = "selected"
	return pattern, nil
}

func (s *ShiftService) FinalizePattern(ctx context.Context, id string) (*model.ShiftPattern, error) {
	pattern, err := s.patternRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pattern == nil {
		return nil, nil
	}

	if err := s.patternRepo.UpdateStatus(ctx, id, "finalized"); err != nil {
		return nil, err
	}

	pattern.Status = "finalized"
	return pattern, nil
}

func (s *ShiftService) CreateEntry(ctx context.Context, req model.CreateShiftEntryRequest) (*model.ShiftEntry, error) {
	return s.entryRepo.Create(ctx, req)
}

func (s *ShiftService) UpdateEntry(ctx context.Context, id string, req model.UpdateShiftEntryRequest) (*model.ShiftEntry, *model.EntryValidation, error) {
	entry, err := s.entryRepo.Update(ctx, id, req)
	if err != nil {
		return nil, nil, err
	}
	if entry == nil {
		return nil, nil, nil
	}

	validation := &model.EntryValidation{
		IsValid:  true,
		Warnings: []model.ValidationWarning{},
	}

	return entry, validation, nil
}

func (s *ShiftService) DeleteEntry(ctx context.Context, id string) error {
	return s.entryRepo.Delete(ctx, id)
}

func (s *ShiftService) computeSummary(ctx context.Context, patternID string) (*model.PatternSummary, error) {
	entries, err := s.entryRepo.ListByPatternID(ctx, patternID)
	if err != nil {
		return nil, err
	}

	staffHours := make(map[string]float64)
	for _, e := range entries {
		hours := computeWorkHours(e.StartTime, e.EndTime, e.BreakMinutes)
		name := e.StaffName
		if name == "" {
			name = e.StaffID
		}
		staffHours[name] += hours
	}

	return &model.PatternSummary{
		TotalEntries: len(entries),
		StaffHours:   staffHours,
	}, nil
}

func computeWorkHours(startTime, endTime string, breakMinutes int) float64 {
	// Parse HH:MM format
	var sh, sm, eh, em int
	fmt.Sscanf(startTime, "%d:%d", &sh, &sm)
	fmt.Sscanf(endTime, "%d:%d", &eh, &em)

	totalMinutes := (eh*60 + em) - (sh*60 + sm) - breakMinutes
	if totalMinutes < 0 {
		totalMinutes = 0
	}
	return float64(totalMinutes) / 60.0
}
