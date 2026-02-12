package service

import (
	"context"
	"errors"

	"shift-app/internal/model"
	"shift-app/internal/repository"
)

type StaffService struct {
	repo *repository.StaffRepository
}

func NewStaffService(repo *repository.StaffRepository) *StaffService {
	return &StaffService{repo: repo}
}

func (s *StaffService) List(ctx context.Context, isActive *bool) ([]model.Staff, error) {
	staffs, err := s.repo.List(ctx, isActive)
	if err != nil {
		return nil, err
	}
	if staffs == nil {
		staffs = []model.Staff{}
	}
	return staffs, nil
}

func (s *StaffService) GetByID(ctx context.Context, id string) (*model.Staff, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *StaffService) Create(ctx context.Context, req model.CreateStaffRequest) (*model.Staff, error) {
	if req.Name == "" {
		return nil, errors.New("名前は必須です")
	}
	if len(req.Name) > 100 {
		return nil, errors.New("名前は100文字以内で入力してください")
	}
	if req.Role == "" {
		return nil, errors.New("役割は必須です")
	}
	validRoles := map[string]bool{"kitchen": true, "hall": true, "both": true}
	if !validRoles[req.Role] {
		return nil, errors.New("役割は kitchen, hall, both のいずれかで指定してください")
	}
	if req.EmploymentType == "" {
		return nil, errors.New("雇用形態は必須です")
	}
	validEmpTypes := map[string]bool{"full_time": true, "part_time": true}
	if !validEmpTypes[req.EmploymentType] {
		return nil, errors.New("雇用形態は full_time, part_time のいずれかで指定してください")
	}
	return s.repo.Create(ctx, req)
}

func (s *StaffService) Update(ctx context.Context, id string, req model.UpdateStaffRequest) (*model.Staff, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *StaffService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
