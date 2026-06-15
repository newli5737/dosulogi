package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/dosu-logi/logistics-erp/internal/sales/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OpportunityRepo struct{ db *pgxpool.Pool }

func NewOpportunityRepo(db *pgxpool.Pool) *OpportunityRepo { return &OpportunityRepo{db: db} }

func (r *OpportunityRepo) Create(ctx context.Context, o *domain.Opportunity) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO opportunities (customer_id, title, stage, value, currency, expected_close, assigned_to, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at, updated_at`,
		o.CustomerID, o.Title, o.Stage, o.Value, o.Currency, o.ExpectedClose, o.AssignedTo, o.CreatedBy,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}

func (r *OpportunityRepo) List(ctx context.Context, f domain.OpportunityFilter) ([]domain.Opportunity, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1
	if f.Stage != "" {
		where += fmt.Sprintf(" AND stage = $%d", n)
		args = append(args, f.Stage)
		n++
	}
	if f.AssignedTo != "" {
		where += fmt.Sprintf(" AND assigned_to = $%d", n)
		args = append(args, f.AssignedTo)
		n++
	}
	if f.Role == "sales_rep" {
		where += fmt.Sprintf(" AND assigned_to = $%d", n)
		args = append(args, f.UserID)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM opportunities "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT id, customer_id, title, stage, value, currency, expected_close, assigned_to, lost_reason, created_by, created_at, updated_at
		FROM opportunities %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []domain.Opportunity
	for rows.Next() {
		var o domain.Opportunity
		_ = rows.Scan(&o.ID, &o.CustomerID, &o.Title, &o.Stage, &o.Value, &o.Currency, &o.ExpectedClose,
			&o.AssignedTo, &o.LostReason, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt)
		list = append(list, o)
	}
	return list, total, nil
}

func (r *OpportunityRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Opportunity, error) {
	o := &domain.Opportunity{}
	err := r.db.QueryRow(ctx, `
		SELECT id, customer_id, title, stage, value, currency, expected_close, assigned_to, lost_reason, created_by, created_at, updated_at
		FROM opportunities WHERE id=$1`, id,
	).Scan(&o.ID, &o.CustomerID, &o.Title, &o.Stage, &o.Value, &o.Currency, &o.ExpectedClose,
		&o.AssignedTo, &o.LostReason, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, applicationErrNotFound
	}
	return o, err
}

func (r *OpportunityRepo) Update(ctx context.Context, o *domain.Opportunity) error {
	_, err := r.db.Exec(ctx, `
		UPDATE opportunities SET title=$1, stage=$2, value=$3, currency=$4, expected_close=$5, assigned_to=$6, lost_reason=$7, updated_at=now()
		WHERE id=$8`, o.Title, o.Stage, o.Value, o.Currency, o.ExpectedClose, o.AssignedTo, o.LostReason, o.ID)
	return err
}

func (r *OpportunityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM opportunities WHERE id=$1`, id)
	return err
}

func (r *OpportunityRepo) InsertStageHistory(ctx context.Context, oppID uuid.UUID, fromStage, toStage string, changedBy *uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO opportunity_stage_history (opportunity_id, from_stage, to_stage, changed_by)
		VALUES ($1, $2, $3, $4)`, oppID, fromStage, toStage, changedBy)
	return err
}

var applicationErrNotFound = errors.New("not found")
