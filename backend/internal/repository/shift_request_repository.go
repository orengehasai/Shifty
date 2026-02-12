package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

type ShiftRequestRepository struct {
	db *pgxpool.Pool
}

func NewShiftRequestRepository(db *pgxpool.Pool) *ShiftRequestRepository {
	return &ShiftRequestRepository{db: db}
}

func (r *ShiftRequestRepository) List(ctx context.Context, yearMonth string, staffID *string) ([]model.ShiftRequest, error) {
	query := `SELECT sr.id, sr.staff_id, s.name, sr.year_month, sr.date::text, sr.start_time::text, sr.end_time::text, sr.request_type, sr.note, sr.created_at, sr.updated_at
		FROM shift_requests sr
		JOIN staffs s ON s.id = sr.staff_id
		WHERE sr.year_month = $1`
	args := []interface{}{yearMonth}

	if staffID != nil {
		query += ` AND sr.staff_id = $2`
		args = append(args, *staffID)
	}
	query += ` ORDER BY sr.date ASC, s.name ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []model.ShiftRequest
	for rows.Next() {
		var sr model.ShiftRequest
		if err := rows.Scan(&sr.ID, &sr.StaffID, &sr.StaffName, &sr.YearMonth, &sr.Date, &sr.StartTime, &sr.EndTime, &sr.RequestType, &sr.Note, &sr.CreatedAt, &sr.UpdatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, sr)
	}
	return requests, rows.Err()
}

func (r *ShiftRequestRepository) GetByID(ctx context.Context, id string) (*model.ShiftRequest, error) {
	var sr model.ShiftRequest
	err := r.db.QueryRow(ctx,
		`SELECT sr.id, sr.staff_id, s.name, sr.year_month, sr.date::text, sr.start_time::text, sr.end_time::text, sr.request_type, sr.note, sr.created_at, sr.updated_at
		 FROM shift_requests sr
		 JOIN staffs s ON s.id = sr.staff_id
		 WHERE sr.id = $1`, id,
	).Scan(&sr.ID, &sr.StaffID, &sr.StaffName, &sr.YearMonth, &sr.Date, &sr.StartTime, &sr.EndTime, &sr.RequestType, &sr.Note, &sr.CreatedAt, &sr.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &sr, nil
}

func (r *ShiftRequestRepository) Create(ctx context.Context, req model.CreateShiftRequestRequest) (*model.ShiftRequest, error) {
	var sr model.ShiftRequest
	err := r.db.QueryRow(ctx,
		`INSERT INTO shift_requests (staff_id, year_month, date, start_time, end_time, request_type, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, staff_id, year_month, date::text, start_time::text, end_time::text, request_type, note, created_at, updated_at`,
		req.StaffID, req.YearMonth, req.Date, req.StartTime, req.EndTime, req.RequestType, req.Note,
	).Scan(&sr.ID, &sr.StaffID, &sr.YearMonth, &sr.Date, &sr.StartTime, &sr.EndTime, &sr.RequestType, &sr.Note, &sr.CreatedAt, &sr.UpdatedAt)
	if err != nil {
		return nil, err
	}

	var staffName string
	_ = r.db.QueryRow(ctx, `SELECT name FROM staffs WHERE id = $1`, sr.StaffID).Scan(&staffName)
	sr.StaffName = staffName

	return &sr, nil
}

func (r *ShiftRequestRepository) Update(ctx context.Context, id string, req model.CreateShiftRequestRequest) (*model.ShiftRequest, error) {
	var sr model.ShiftRequest
	err := r.db.QueryRow(ctx,
		`UPDATE shift_requests SET start_time=$1, end_time=$2, request_type=$3, note=$4, updated_at=NOW()
		 WHERE id=$5
		 RETURNING id, staff_id, year_month, date::text, start_time::text, end_time::text, request_type, note, created_at, updated_at`,
		req.StartTime, req.EndTime, req.RequestType, req.Note, id,
	).Scan(&sr.ID, &sr.StaffID, &sr.YearMonth, &sr.Date, &sr.StartTime, &sr.EndTime, &sr.RequestType, &sr.Note, &sr.CreatedAt, &sr.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var staffName string
	_ = r.db.QueryRow(ctx, `SELECT name FROM staffs WHERE id = $1`, sr.StaffID).Scan(&staffName)
	sr.StaffName = staffName

	return &sr, nil
}

func (r *ShiftRequestRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM shift_requests WHERE id = $1`, id)
	return err
}

func (r *ShiftRequestRepository) CountDistinctStaffByYearMonth(ctx context.Context, yearMonth string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT staff_id) FROM shift_requests WHERE year_month = $1`, yearMonth,
	).Scan(&count)
	return count, err
}
