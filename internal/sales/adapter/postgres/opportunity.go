package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/dosu-logi/logistics-erp/internal/platform/codegen"
	"github.com/dosu-logi/logistics-erp/internal/sales/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OpportunityRepo struct{ db *pgxpool.Pool }

func NewOpportunityRepo(db *pgxpool.Pool) *OpportunityRepo { return &OpportunityRepo{db: db} }

func (r *OpportunityRepo) NextCode(ctx context.Context) (string, error) {
	return codegen.Next(ctx, r.db, "opportunities", "OPP", false)
}

func (r *OpportunityRepo) Create(ctx context.Context, o *domain.Opportunity) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO opportunities (code, customer_id, title, stage, value, currency, expected_close, assigned_to, created_by, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, created_at, updated_at`,
		o.Code, o.CustomerID, o.Title, o.Stage, o.Value, o.Currency, o.ExpectedClose, o.AssignedTo, o.CreatedBy, o.Note,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}

func (r *OpportunityRepo) List(ctx context.Context, f domain.OpportunityFilter) ([]domain.Opportunity, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1
	if f.Stage != "" {
		where += fmt.Sprintf(" AND o.stage = $%d", n)
		args = append(args, f.Stage)
		n++
	}
	if f.AssignedTo != "" {
		where += fmt.Sprintf(" AND o.assigned_to = $%d", n)
		args = append(args, f.AssignedTo)
		n++
	}
	if f.Role == "sales_rep" {
		where += fmt.Sprintf(" AND o.assigned_to = $%d", n)
		args = append(args, f.UserID)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM opportunities o "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT o.id, o.code, o.customer_id, o.title, o.stage, o.value, o.currency, o.expected_close, o.assigned_to, o.lost_reason, o.note, o.created_by, o.created_at, o.updated_at,
		c.id, c.name, c.code
		FROM opportunities o JOIN customers c ON c.id = o.customer_id %s ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []domain.Opportunity
	for rows.Next() {
		var o domain.Opportunity
		var cid uuid.UUID
		var cname, ccode string
		_ = rows.Scan(&o.ID, &o.Code, &o.CustomerID, &o.Title, &o.Stage, &o.Value, &o.Currency, &o.ExpectedClose,
			&o.AssignedTo, &o.LostReason, &o.Note, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt, &cid, &cname, &ccode)
		o.Customer = &struct {
			ID   uuid.UUID `json:"id"`
			Name string    `json:"name"`
			Code string    `json:"code"`
		}{ID: cid, Name: cname, Code: ccode}
		list = append(list, o)
	}
	if list == nil {
		list = []domain.Opportunity{}
	}
	return list, total, nil
}

func (r *OpportunityRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Opportunity, error) {
	o := &domain.Opportunity{}
	var cid uuid.UUID
	var cname, ccode string
	err := r.db.QueryRow(ctx, `
		SELECT o.id, o.code, o.customer_id, o.title, o.stage, o.value, o.currency, o.expected_close, o.assigned_to, o.lost_reason, o.note, o.created_by, o.created_at, o.updated_at,
		c.id, c.name, c.code
		FROM opportunities o JOIN customers c ON c.id = o.customer_id WHERE o.id=$1`, id,
	).Scan(&o.ID, &o.Code, &o.CustomerID, &o.Title, &o.Stage, &o.Value, &o.Currency, &o.ExpectedClose,
		&o.AssignedTo, &o.LostReason, &o.Note, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt, &cid, &cname, &ccode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, applicationErrNotFound
	}
	if err != nil {
		return nil, err
	}
	o.Customer = &struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Code string    `json:"code"`
	}{ID: cid, Name: cname, Code: ccode}
	ids, _ := r.GetShipmentIDs(ctx, id)
	o.ShipmentIDs = ids
	return o, nil
}

func (r *OpportunityRepo) Update(ctx context.Context, o *domain.Opportunity) error {
	_, err := r.db.Exec(ctx, `
		UPDATE opportunities SET title=$1, stage=$2, value=$3, currency=$4, expected_close=$5, assigned_to=$6, lost_reason=$7, note=$8, updated_at=now()
		WHERE id=$9`, o.Title, o.Stage, o.Value, o.Currency, o.ExpectedClose, o.AssignedTo, o.LostReason, o.Note, o.ID)
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

func (r *OpportunityRepo) ListStageHistory(ctx context.Context, oppID uuid.UUID) ([]domain.StageHistoryEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT h.id, h.from_stage, h.to_stage, h.note, h.changed_by, h.changed_at, u.full_name
		FROM opportunity_stage_history h LEFT JOIN users u ON u.id = h.changed_by
		WHERE h.opportunity_id = $1 ORDER BY h.changed_at DESC`, oppID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.StageHistoryEntry
	for rows.Next() {
		var e domain.StageHistoryEntry
		_ = rows.Scan(&e.ID, &e.FromStage, &e.ToStage, &e.Note, &e.ChangedBy, &e.ChangedAt, &e.ChangerName)
		list = append(list, e)
	}
	if list == nil {
		list = []domain.StageHistoryEntry{}
	}
	return list, nil
}

func (r *OpportunityRepo) GetShipmentIDs(ctx context.Context, oppID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT shipment_id FROM opportunity_shipments WHERE opportunity_id=$1`, oppID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		_ = rows.Scan(&id)
		ids = append(ids, id)
	}
	if ids == nil {
		ids = []uuid.UUID{}
	}
	return ids, nil
}

func (r *OpportunityRepo) SetShipments(ctx context.Context, oppID uuid.UUID, ids []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM opportunity_shipments WHERE opportunity_id=$1`, oppID); err != nil {
		return err
	}
	for _, sid := range ids {
		if _, err := tx.Exec(ctx, `INSERT INTO opportunity_shipments (opportunity_id, shipment_id) VALUES ($1,$2)`, oppID, sid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

var applicationErrNotFound = errors.New("not found")
