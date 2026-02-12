package service

import (
	"context"
	"errors"

	"shift-app/internal/model"
	"shift-app/internal/repository"
)

type ConstraintService struct {
	repo *repository.ConstraintRepository
}

func NewConstraintService(repo *repository.ConstraintRepository) *ConstraintService {
	return &ConstraintService{repo: repo}
}

func (s *ConstraintService) List(ctx context.Context, isActive *bool, cType *string, category *string) ([]model.Constraint, error) {
	constraints, err := s.repo.List(ctx, isActive, cType, category)
	if err != nil {
		return nil, err
	}
	if constraints == nil {
		constraints = []model.Constraint{}
	}
	return constraints, nil
}

func (s *ConstraintService) GetByID(ctx context.Context, id string) (*model.Constraint, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ConstraintService) Create(ctx context.Context, req model.CreateConstraintRequest) (*model.Constraint, error) {
	if req.Name == "" {
		return nil, errors.New("名前は必須です")
	}
	if len(req.Name) > 200 {
		return nil, errors.New("名前は200文字以内で入力してください")
	}
	if req.Type == "" {
		return nil, errors.New("type は必須です")
	}
	validTypes := map[string]bool{"hard": true, "soft": true}
	if !validTypes[req.Type] {
		return nil, errors.New("type は hard, soft のいずれかで指定してください")
	}
	if req.Category == "" {
		return nil, errors.New("category は必須です")
	}
	validCategories := map[string]bool{
		"min_staff": true, "max_staff": true, "max_consecutive_days": true,
		"monthly_hours": true, "fixed_day_off": true, "staff_compatibility": true, "rest_hours": true,
	}
	if !validCategories[req.Category] {
		return nil, errors.New("無効な category です")
	}
	return s.repo.Create(ctx, req)
}

func (s *ConstraintService) Update(ctx context.Context, id string, req model.UpdateConstraintRequest) (*model.Constraint, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *ConstraintService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
