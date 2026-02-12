package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

type ShiftPatternRepository struct {
	db *pgxpool.Pool
}

func NewShiftPatternRepository(db *pgxpool.Pool) *ShiftPatternRepository {
	return &ShiftPatternRepository{db: db}
}

func (r *ShiftPatternRepository) ListByYearMonth(ctx context.Context, yearMonth string) ([]model.ShiftPattern, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, year_month, status, reasoning, score, constraint_violations, created_at, updated_at
		 FROM shift_patterns WHERE year_month = $1 ORDER BY created_at ASC`, yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []model.ShiftPattern
	for rows.Next() {
		var p model.ShiftPattern
		var violationsJSON []byte
		if err := rows.Scan(&p.ID, &p.YearMonth, &p.Status, &p.Reasoning, &p.Score, &violationsJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if violationsJSON != nil {
			_ = json.Unmarshal(violationsJSON, &p.ConstraintViolations)
		}
		if p.ConstraintViolations == nil {
			p.ConstraintViolations = []model.ConstraintViolation{}
		}
		patterns = append(patterns, p)
	}
	return patterns, rows.Err()
}

func (r *ShiftPatternRepository) GetByID(ctx context.Context, id string) (*model.ShiftPattern, error) {
	var p model.ShiftPattern
	var violationsJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, year_month, status, reasoning, score, constraint_violations, created_at, updated_at
		 FROM shift_patterns WHERE id = $1`, id,
	).Scan(&p.ID, &p.YearMonth, &p.Status, &p.Reasoning, &p.Score, &violationsJSON, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if violationsJSON != nil {
		_ = json.Unmarshal(violationsJSON, &p.ConstraintViolations)
	}
	if p.ConstraintViolations == nil {
		p.ConstraintViolations = []model.ConstraintViolation{}
	}
	return &p, nil
}

func (r *ShiftPatternRepository) Create(ctx context.Context, yearMonth string, reasoning string, score float64, violations []model.ConstraintViolation) (*model.ShiftPattern, error) {
	violationsJSON, _ := json.Marshal(violations)

	var p model.ShiftPattern
	var violBytes []byte
	err := r.db.QueryRow(ctx,
		`INSERT INTO shift_patterns (year_month, reasoning, score, constraint_violations)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, year_month, status, reasoning, score, constraint_violations, created_at, updated_at`,
		yearMonth, reasoning, score, violationsJSON,
	).Scan(&p.ID, &p.YearMonth, &p.Status, &p.Reasoning, &p.Score, &violBytes, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if violBytes != nil {
		_ = json.Unmarshal(violBytes, &p.ConstraintViolations)
	}
	if p.ConstraintViolations == nil {
		p.ConstraintViolations = []model.ConstraintViolation{}
	}
	return &p, nil
}

func (r *ShiftPatternRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE shift_patterns SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

// ResetOtherPatterns sets all other patterns for the same year_month to 'draft'
func (r *ShiftPatternRepository) ResetOtherPatterns(ctx context.Context, id string, yearMonth string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE shift_patterns SET status = 'draft', updated_at = NOW() WHERE year_month = $1 AND id != $2 AND status != 'finalized'`,
		yearMonth, id)
	return err
}

func (r *ShiftPatternRepository) GetLatestStatusByYearMonth(ctx context.Context, yearMonth string) (string, error) {
	var status string
	err := r.db.QueryRow(ctx,
		`SELECT status FROM shift_patterns WHERE year_month = $1 ORDER BY
		 CASE status
		   WHEN 'finalized' THEN 1
		   WHEN 'selected' THEN 2
		   WHEN 'draft' THEN 3
		 END ASC
		 LIMIT 1`, yearMonth,
	).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return status, nil
}
