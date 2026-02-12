package service

import (
	"context"
	"errors"

	"shift-app/internal/model"
	"shift-app/internal/repository"
)

type ShiftRequestService struct {
	repo *repository.ShiftRequestRepository
}

func NewShiftRequestService(repo *repository.ShiftRequestRepository) *ShiftRequestService {
	return &ShiftRequestService{repo: repo}
}

func (s *ShiftRequestService) List(ctx context.Context, yearMonth string, staffID *string) ([]model.ShiftRequest, error) {
	if yearMonth == "" {
		return nil, errors.New("year_month は必須です")
	}
	requests, err := s.repo.List(ctx, yearMonth, staffID)
	if err != nil {
		return nil, err
	}
	if requests == nil {
		requests = []model.ShiftRequest{}
	}
	return requests, nil
}

func (s *ShiftRequestService) Create(ctx context.Context, req model.CreateShiftRequestRequest) (*model.ShiftRequest, error) {
	if req.StaffID == "" {
		return nil, errors.New("staff_id は必須です")
	}
	if req.YearMonth == "" {
		return nil, errors.New("year_month は必須です")
	}
	if len(req.YearMonth) != 7 {
		return nil, errors.New("year_month は YYYY-MM 形式で指定してください")
	}
	if req.Date == "" {
		return nil, errors.New("date は必須です")
	}
	if req.RequestType == "" {
		return nil, errors.New("request_type は必須です")
	}
	validTypes := map[string]bool{"available": true, "unavailable": true, "preferred": true}
	if !validTypes[req.RequestType] {
		return nil, errors.New("request_type は available, unavailable, preferred のいずれかで指定してください")
	}
	return s.repo.Create(ctx, req)
}

func (s *ShiftRequestService) BatchCreate(ctx context.Context, req model.BatchShiftRequestRequest) ([]model.ShiftRequest, error) {
	if len(req.Requests) == 0 {
		return []model.ShiftRequest{}, nil
	}
	if len(req.Requests) > 100 {
		return nil, errors.New("一括登録は100件以内で指定してください")
	}
	var results []model.ShiftRequest
	for _, r := range req.Requests {
		result, err := s.repo.Create(ctx, r)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	if results == nil {
		results = []model.ShiftRequest{}
	}
	return results, nil
}

func (s *ShiftRequestService) Update(ctx context.Context, id string, req model.CreateShiftRequestRequest) (*model.ShiftRequest, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *ShiftRequestService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
