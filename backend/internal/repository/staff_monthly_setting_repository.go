package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

type StaffMonthlySettingRepository struct {
	db *pgxpool.Pool
}

func NewStaffMonthlySettingRepository(db *pgxpool.Pool) *StaffMonthlySettingRepository {
	return &StaffMonthlySettingRepository{db: db}
}

func (r *StaffMonthlySettingRepository) List(ctx context.Context, yearMonth string, staffID *string) ([]model.StaffMonthlySetting, error) {
	query := `SELECT sms.id, sms.staff_id, s.name, sms.year_month, sms.min_preferred_hours, sms.max_preferred_hours, sms.note, sms.created_at, sms.updated_at
		FROM staff_monthly_settings sms
		JOIN staffs s ON s.id = sms.staff_id
		WHERE sms.year_month = $1`
	args := []interface{}{yearMonth}

	if staffID != nil {
		query += ` AND sms.staff_id = $2`
		args = append(args, *staffID)
	}
	query += ` ORDER BY s.name ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []model.StaffMonthlySetting
	for rows.Next() {
		var s model.StaffMonthlySetting
		if err := rows.Scan(&s.ID, &s.StaffID, &s.StaffName, &s.YearMonth, &s.MinPreferredHours, &s.MaxPreferredHours, &s.Note, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}

func (r *StaffMonthlySettingRepository) GetByID(ctx context.Context, id string) (*model.StaffMonthlySetting, error) {
	var s model.StaffMonthlySetting
	err := r.db.QueryRow(ctx,
		`SELECT sms.id, sms.staff_id, st.name, sms.year_month, sms.min_preferred_hours, sms.max_preferred_hours, sms.note, sms.created_at, sms.updated_at
		 FROM staff_monthly_settings sms
		 JOIN staffs st ON st.id = sms.staff_id
		 WHERE sms.id = $1`, id,
	).Scan(&s.ID, &s.StaffID, &s.StaffName, &s.YearMonth, &s.MinPreferredHours, &s.MaxPreferredHours, &s.Note, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *StaffMonthlySettingRepository) Upsert(ctx context.Context, req model.CreateStaffMonthlySettingRequest) (*model.StaffMonthlySetting, error) {
	var s model.StaffMonthlySetting
	err := r.db.QueryRow(ctx,
		`INSERT INTO staff_monthly_settings (staff_id, year_month, min_preferred_hours, max_preferred_hours, note)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (staff_id, year_month)
		 DO UPDATE SET min_preferred_hours = EXCLUDED.min_preferred_hours,
		               max_preferred_hours = EXCLUDED.max_preferred_hours,
		               note = EXCLUDED.note,
		               updated_at = NOW()
		 RETURNING id, staff_id, year_month, min_preferred_hours, max_preferred_hours, note, created_at, updated_at`,
		req.StaffID, req.YearMonth, req.MinPreferredHours, req.MaxPreferredHours, req.Note,
	).Scan(&s.ID, &s.StaffID, &s.YearMonth, &s.MinPreferredHours, &s.MaxPreferredHours, &s.Note, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Fetch staff name
	var staffName string
	_ = r.db.QueryRow(ctx, `SELECT name FROM staffs WHERE id = $1`, s.StaffID).Scan(&staffName)
	s.StaffName = staffName

	return &s, nil
}

func (r *StaffMonthlySettingRepository) Update(ctx context.Context, id string, req model.CreateStaffMonthlySettingRequest) (*model.StaffMonthlySetting, error) {
	var s model.StaffMonthlySetting
	err := r.db.QueryRow(ctx,
		`UPDATE staff_monthly_settings SET min_preferred_hours=$1, max_preferred_hours=$2, note=$3, updated_at=NOW()
		 WHERE id=$4
		 RETURNING id, staff_id, year_month, min_preferred_hours, max_preferred_hours, note, created_at, updated_at`,
		req.MinPreferredHours, req.MaxPreferredHours, req.Note, id,
	).Scan(&s.ID, &s.StaffID, &s.YearMonth, &s.MinPreferredHours, &s.MaxPreferredHours, &s.Note, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var staffName string
	_ = r.db.QueryRow(ctx, `SELECT name FROM staffs WHERE id = $1`, s.StaffID).Scan(&staffName)
	s.StaffName = staffName

	return &s, nil
}

func (r *StaffMonthlySettingRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM staff_monthly_settings WHERE id = $1`, id)
	return err
}

func (r *StaffMonthlySettingRepository) CountByYearMonth(ctx context.Context, yearMonth string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT staff_id) FROM staff_monthly_settings WHERE year_month = $1`, yearMonth,
	).Scan(&count)
	return count, err
}
