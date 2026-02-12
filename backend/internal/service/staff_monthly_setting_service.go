package service

import (
	"context"
	"errors"

	"shift-app/internal/model"
	"shift-app/internal/repository"
)

type StaffMonthlySettingService struct {
	repo *repository.StaffMonthlySettingRepository
}

func NewStaffMonthlySettingService(repo *repository.StaffMonthlySettingRepository) *StaffMonthlySettingService {
	return &StaffMonthlySettingService{repo: repo}
}

func (s *StaffMonthlySettingService) List(ctx context.Context, yearMonth string, staffID *string) ([]model.StaffMonthlySetting, error) {
	if yearMonth == "" {
		return nil, errors.New("year_month は必須です")
	}
	settings, err := s.repo.List(ctx, yearMonth, staffID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings = []model.StaffMonthlySetting{}
	}
	return settings, nil
}

func (s *StaffMonthlySettingService) Create(ctx context.Context, req model.CreateStaffMonthlySettingRequest) (*model.StaffMonthlySetting, error) {
	if req.StaffID == "" {
		return nil, errors.New("staff_id は必須です")
	}
	if req.YearMonth == "" {
		return nil, errors.New("year_month は必須です")
	}
	if req.MinPreferredHours < 0 || req.MaxPreferredHours < 0 {
		return nil, errors.New("希望時間は0以上で指定してください")
	}
	if req.MinPreferredHours > req.MaxPreferredHours {
		return nil, errors.New("最小希望時間は最大希望時間以下にしてください")
	}
	if req.MaxPreferredHours > 744 {
		return nil, errors.New("最大希望時間が大きすぎます")
	}
	return s.repo.Upsert(ctx, req)
}

func (s *StaffMonthlySettingService) BatchCreate(ctx context.Context, req model.BatchStaffMonthlySettingRequest) ([]model.StaffMonthlySetting, error) {
	if len(req.Settings) == 0 {
		return []model.StaffMonthlySetting{}, nil
	}
	if len(req.Settings) > 100 {
		return nil, errors.New("一括登録は100件以内で指定してください")
	}
	var results []model.StaffMonthlySetting
	for _, setting := range req.Settings {
		result, err := s.repo.Upsert(ctx, setting)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	if results == nil {
		results = []model.StaffMonthlySetting{}
	}
	return results, nil
}

func (s *StaffMonthlySettingService) Update(ctx context.Context, id string, req model.CreateStaffMonthlySettingRequest) (*model.StaffMonthlySetting, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *StaffMonthlySettingService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
