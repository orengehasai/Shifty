package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

type GenerationJobRepository struct {
	db *pgxpool.Pool
}

func NewGenerationJobRepository(db *pgxpool.Pool) *GenerationJobRepository {
	return &GenerationJobRepository{db: db}
}

func (r *GenerationJobRepository) Create(ctx context.Context, yearMonth string, patternCount int) (*model.GenerationJob, error) {
	var j model.GenerationJob
	err := r.db.QueryRow(ctx,
		`INSERT INTO generation_jobs (year_month, pattern_count) VALUES ($1, $2)
		 RETURNING id, year_month, status, pattern_count, progress, status_message, error_message, started_at, completed_at, created_at`,
		yearMonth, patternCount,
	).Scan(&j.ID, &j.YearMonth, &j.Status, &j.PatternCount, &j.Progress, &j.StatusMessage, &j.ErrorMessage, &j.StartedAt, &j.CompletedAt, &j.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *GenerationJobRepository) GetByID(ctx context.Context, id string) (*model.GenerationJob, error) {
	var j model.GenerationJob
	err := r.db.QueryRow(ctx,
		`SELECT id, year_month, status, pattern_count, progress, status_message, error_message, started_at, completed_at, created_at
		 FROM generation_jobs WHERE id = $1`, id,
	).Scan(&j.ID, &j.YearMonth, &j.Status, &j.PatternCount, &j.Progress, &j.StatusMessage, &j.ErrorMessage, &j.StartedAt, &j.CompletedAt, &j.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &j, nil
}

func (r *GenerationJobRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE generation_jobs SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *GenerationJobRepository) SetProcessing(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE generation_jobs SET status = 'processing', started_at = $1 WHERE id = $2`, now, id)
	return err
}

func (r *GenerationJobRepository) SetCompleted(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE generation_jobs SET status = 'completed', completed_at = $1 WHERE id = $2`, now, id)
	return err
}

func (r *GenerationJobRepository) SetFailed(ctx context.Context, id string, errMsg string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE generation_jobs SET status = 'failed', error_message = $1, completed_at = $2 WHERE id = $3`, errMsg, now, id)
	return err
}

func (r *GenerationJobRepository) UpdateProgress(ctx context.Context, id string, progress int, statusMessage string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE generation_jobs SET progress = $1, status_message = $2 WHERE id = $3`, progress, statusMessage, id)
	return err
}

func (r *GenerationJobRepository) HasProcessingJob(ctx context.Context, yearMonth string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM generation_jobs WHERE year_month = $1 AND status IN ('pending', 'processing')`, yearMonth,
	).Scan(&count)
	return count > 0, err
}
