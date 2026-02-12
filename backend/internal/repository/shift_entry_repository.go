package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

type ShiftEntryRepository struct {
	db *pgxpool.Pool
}

func NewShiftEntryRepository(db *pgxpool.Pool) *ShiftEntryRepository {
	return &ShiftEntryRepository{db: db}
}

func (r *ShiftEntryRepository) ListByPatternID(ctx context.Context, patternID string) ([]model.ShiftEntry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT se.id, se.pattern_id, se.staff_id, s.name, se.date::text, se.start_time::text, se.end_time::text, se.break_minutes, se.is_manual_edit, se.created_at, se.updated_at
		 FROM shift_entries se
		 JOIN staffs s ON s.id = se.staff_id
		 WHERE se.pattern_id = $1
		 ORDER BY se.date ASC, se.start_time ASC, s.name ASC`, patternID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.ShiftEntry
	for rows.Next() {
		var e model.ShiftEntry
		if err := rows.Scan(&e.ID, &e.PatternID, &e.StaffID, &e.StaffName, &e.Date, &e.StartTime, &e.EndTime, &e.BreakMinutes, &e.IsManualEdit, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *ShiftEntryRepository) GetByID(ctx context.Context, id string) (*model.ShiftEntry, error) {
	var e model.ShiftEntry
	err := r.db.QueryRow(ctx,
		`SELECT se.id, se.pattern_id, se.staff_id, s.name, se.date::text, se.start_time::text, se.end_time::text, se.break_minutes, se.is_manual_edit, se.created_at, se.updated_at
		 FROM shift_entries se
		 JOIN staffs s ON s.id = se.staff_id
		 WHERE se.id = $1`, id,
	).Scan(&e.ID, &e.PatternID, &e.StaffID, &e.StaffName, &e.Date, &e.StartTime, &e.EndTime, &e.BreakMinutes, &e.IsManualEdit, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *ShiftEntryRepository) Create(ctx context.Context, req model.CreateShiftEntryRequest) (*model.ShiftEntry, error) {
	var e model.ShiftEntry
	err := r.db.QueryRow(ctx,
		`INSERT INTO shift_entries (pattern_id, staff_id, date, start_time, end_time, break_minutes, is_manual_edit)
		 VALUES ($1, $2, $3, $4, $5, $6, true)
		 RETURNING id, pattern_id, staff_id, date::text, start_time::text, end_time::text, break_minutes, is_manual_edit, created_at, updated_at`,
		req.PatternID, req.StaffID, req.Date, req.StartTime, req.EndTime, req.BreakMinutes,
	).Scan(&e.ID, &e.PatternID, &e.StaffID, &e.Date, &e.StartTime, &e.EndTime, &e.BreakMinutes, &e.IsManualEdit, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}

	var staffName string
	_ = r.db.QueryRow(ctx, `SELECT name FROM staffs WHERE id = $1`, e.StaffID).Scan(&staffName)
	e.StaffName = staffName

	return &e, nil
}

func (r *ShiftEntryRepository) Update(ctx context.Context, id string, req model.UpdateShiftEntryRequest) (*model.ShiftEntry, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	startTime := current.StartTime
	endTime := current.EndTime
	breakMinutes := current.BreakMinutes

	if req.StartTime != nil {
		startTime = *req.StartTime
	}
	if req.EndTime != nil {
		endTime = *req.EndTime
	}
	if req.BreakMinutes != nil {
		breakMinutes = *req.BreakMinutes
	}

	var e model.ShiftEntry
	err = r.db.QueryRow(ctx,
		`UPDATE shift_entries SET start_time=$1, end_time=$2, break_minutes=$3, is_manual_edit=true, updated_at=NOW()
		 WHERE id=$4
		 RETURNING id, pattern_id, staff_id, date::text, start_time::text, end_time::text, break_minutes, is_manual_edit, created_at, updated_at`,
		startTime, endTime, breakMinutes, id,
	).Scan(&e.ID, &e.PatternID, &e.StaffID, &e.Date, &e.StartTime, &e.EndTime, &e.BreakMinutes, &e.IsManualEdit, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}

	var staffName string
	_ = r.db.QueryRow(ctx, `SELECT name FROM staffs WHERE id = $1`, e.StaffID).Scan(&staffName)
	e.StaffName = staffName

	return &e, nil
}

func (r *ShiftEntryRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM shift_entries WHERE id = $1`, id)
	return err
}

func (r *ShiftEntryRepository) BulkCreate(ctx context.Context, patternID string, entries []model.LLMShiftEntry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, entry := range entries {
		_, err := tx.Exec(ctx,
			`INSERT INTO shift_entries (pattern_id, staff_id, date, start_time, end_time, break_minutes)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			patternID, entry.StaffID, entry.Date, entry.StartTime, entry.EndTime, entry.BreakMinutes)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// CountByPatternDate returns the count of entries for a given date in a pattern
func (r *ShiftEntryRepository) CountByPatternDate(ctx context.Context, patternID string, date string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM shift_entries WHERE pattern_id = $1 AND date = $2`, patternID, date,
	).Scan(&count)
	return count, err
}

// DailyStaffCounts returns date->count for a pattern
func (r *ShiftEntryRepository) DailyStaffCounts(ctx context.Context, patternID string) ([]model.DailyStaffCount, error) {
	rows, err := r.db.Query(ctx,
		`SELECT date::text, COUNT(*) as cnt FROM shift_entries WHERE pattern_id = $1 GROUP BY date ORDER BY date`, patternID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []model.DailyStaffCount
	for rows.Next() {
		var c model.DailyStaffCount
		if err := rows.Scan(&c.Date, &c.Count); err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	return counts, rows.Err()
}
