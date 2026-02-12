package repository

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shift-app/internal/model"
)

type ConstraintRepository struct {
	db *pgxpool.Pool
}

func NewConstraintRepository(db *pgxpool.Pool) *ConstraintRepository {
	return &ConstraintRepository{db: db}
}

func (r *ConstraintRepository) List(ctx context.Context, isActive *bool, cType *string, category *string) ([]model.Constraint, error) {
	query := `SELECT id, name, type, category, config, is_active, priority, created_at, updated_at FROM constraints WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if isActive != nil {
		query += ` AND is_active = $` + itoa(argIdx)
		args = append(args, *isActive)
		argIdx++
	}
	if cType != nil {
		query += ` AND type = $` + itoa(argIdx)
		args = append(args, *cType)
		argIdx++
	}
	if category != nil {
		query += ` AND category = $` + itoa(argIdx)
		args = append(args, *category)
		argIdx++
	}
	query += ` ORDER BY priority DESC, created_at ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var constraints []model.Constraint
	for rows.Next() {
		var c model.Constraint
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Category, &c.Config, &c.IsActive, &c.Priority, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		constraints = append(constraints, c)
	}
	return constraints, rows.Err()
}

func (r *ConstraintRepository) GetByID(ctx context.Context, id string) (*model.Constraint, error) {
	var c model.Constraint
	err := r.db.QueryRow(ctx,
		`SELECT id, name, type, category, config, is_active, priority, created_at, updated_at FROM constraints WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Type, &c.Category, &c.Config, &c.IsActive, &c.Priority, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *ConstraintRepository) Create(ctx context.Context, req model.CreateConstraintRequest) (*model.Constraint, error) {
	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}
	var c model.Constraint
	err := r.db.QueryRow(ctx,
		`INSERT INTO constraints (name, type, category, config, priority)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name, type, category, config, is_active, priority, created_at, updated_at`,
		req.Name, req.Type, req.Category, req.Config, priority,
	).Scan(&c.ID, &c.Name, &c.Type, &c.Category, &c.Config, &c.IsActive, &c.Priority, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ConstraintRepository) Update(ctx context.Context, id string, req model.UpdateConstraintRequest) (*model.Constraint, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	name := current.Name
	cType := current.Type
	category := current.Category
	config := current.Config
	isActive := current.IsActive
	priority := current.Priority

	if req.Name != nil {
		name = *req.Name
	}
	if req.Type != nil {
		cType = *req.Type
	}
	if req.Category != nil {
		category = *req.Category
	}
	if req.Config != nil {
		config = *req.Config
	}
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.Priority != nil {
		priority = req.Priority
	}

	var c model.Constraint
	err = r.db.QueryRow(ctx,
		`UPDATE constraints SET name=$1, type=$2, category=$3, config=$4, is_active=$5, priority=$6, updated_at=NOW()
		 WHERE id=$7
		 RETURNING id, name, type, category, config, is_active, priority, created_at, updated_at`,
		name, cType, category, config, isActive, priority, id,
	).Scan(&c.ID, &c.Name, &c.Type, &c.Category, &c.Config, &c.IsActive, &c.Priority, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ConstraintRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM constraints WHERE id = $1`, id)
	return err
}

func (r *ConstraintRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM constraints WHERE is_active = true`).Scan(&count)
	return count, err
}

// itoa converts int to string for dynamic query building
func itoa(i int) string {
	return strconv.Itoa(i)
}
