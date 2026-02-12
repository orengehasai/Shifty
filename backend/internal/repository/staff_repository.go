package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

type StaffRepository struct {
	db *pgxpool.Pool
}

func NewStaffRepository(db *pgxpool.Pool) *StaffRepository {
	return &StaffRepository{db: db}
}

func (r *StaffRepository) List(ctx context.Context, isActive *bool) ([]model.Staff, error) {
	query := `SELECT id, name, role, employment_type, is_active, created_at, updated_at FROM staffs`
	args := []interface{}{}
	if isActive != nil {
		query += ` WHERE is_active = $1`
		args = append(args, *isActive)
	}
	query += ` ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffs []model.Staff
	for rows.Next() {
		var s model.Staff
		if err := rows.Scan(&s.ID, &s.Name, &s.Role, &s.EmploymentType, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		staffs = append(staffs, s)
	}
	return staffs, rows.Err()
}

func (r *StaffRepository) GetByID(ctx context.Context, id string) (*model.Staff, error) {
	var s model.Staff
	err := r.db.QueryRow(ctx,
		`SELECT id, name, role, employment_type, is_active, created_at, updated_at FROM staffs WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.Name, &s.Role, &s.EmploymentType, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *StaffRepository) Create(ctx context.Context, req model.CreateStaffRequest) (*model.Staff, error) {
	var s model.Staff
	err := r.db.QueryRow(ctx,
		`INSERT INTO staffs (name, role, employment_type) VALUES ($1, $2, $3)
		 RETURNING id, name, role, employment_type, is_active, created_at, updated_at`,
		req.Name, req.Role, req.EmploymentType,
	).Scan(&s.ID, &s.Name, &s.Role, &s.EmploymentType, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StaffRepository) Update(ctx context.Context, id string, req model.UpdateStaffRequest) (*model.Staff, error) {
	// Build dynamic update
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	name := current.Name
	role := current.Role
	empType := current.EmploymentType
	isActive := current.IsActive

	if req.Name != nil {
		name = *req.Name
	}
	if req.Role != nil {
		role = *req.Role
	}
	if req.EmploymentType != nil {
		empType = *req.EmploymentType
	}
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var s model.Staff
	err = r.db.QueryRow(ctx,
		`UPDATE staffs SET name=$1, role=$2, employment_type=$3, is_active=$4, updated_at=NOW()
		 WHERE id=$5
		 RETURNING id, name, role, employment_type, is_active, created_at, updated_at`,
		name, role, empType, isActive, id,
	).Scan(&s.ID, &s.Name, &s.Role, &s.EmploymentType, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StaffRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE staffs SET is_active = false, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *StaffRepository) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM staffs`).Scan(&count)
	return count, err
}

func (r *StaffRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM staffs WHERE is_active = true`).Scan(&count)
	return count, err
}
