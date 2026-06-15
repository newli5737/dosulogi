package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/crm/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InteractionRepo struct{ db *pgxpool.Pool }

func NewInteractionRepo(db *pgxpool.Pool) *InteractionRepo { return &InteractionRepo{db: db} }

func (r *InteractionRepo) List(ctx context.Context, f domain.InteractionFilter) ([]domain.Interaction, int, error) {
	where := "WHERE i.customer_id = $1"
	args := []any{f.CustomerID}
	n := 2
	if f.Channel != "" {
		where += fmt.Sprintf(" AND i.channel = $%d", n)
		args = append(args, f.Channel)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM interactions i "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT i.id, i.customer_id, i.channel, i.direction, i.summary, i.occurred_at, i.created_at, u.id, u.full_name
		FROM interactions i LEFT JOIN users u ON u.id = i.created_by %s ORDER BY i.occurred_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []domain.Interaction
	for rows.Next() {
		var it domain.Interaction
		var uid *uuid.UUID
		var uname *string
		_ = rows.Scan(&it.ID, &it.CustomerID, &it.Channel, &it.Direction, &it.Summary, &it.OccurredAt, &it.CreatedAt, &uid, &uname)
		if uid != nil && uname != nil {
			it.CreatedBy = &domain.UserBrief{ID: *uid, FullName: *uname}
		}
		list = append(list, it)
	}
	if list == nil {
		list = []domain.Interaction{}
	}
	return list, total, nil
}

func (r *InteractionRepo) Create(ctx context.Context, i *domain.Interaction, createdBy uuid.UUID) error {
	at := i.OccurredAt
	if at.IsZero() {
		at = time.Now()
	}
	return r.db.QueryRow(ctx, `
		INSERT INTO interactions (customer_id, channel, direction, summary, occurred_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, occurred_at, created_at`,
		i.CustomerID, i.Channel, i.Direction, i.Summary, at, createdBy,
	).Scan(&i.ID, &i.OccurredAt, &i.CreatedAt)
}

func (r *InteractionRepo) TouchLastContact(ctx context.Context, customerID uuid.UUID, at time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE customers SET last_contact_at=$1, updated_at=now() WHERE id=$2`, at, customerID)
	return err
}
