package sales

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) NextCode(ctx context.Context, prefix string, table string) (string, error) {
	year := time.Now().Year()
	pattern := fmt.Sprintf("%s-%d-%%", prefix, year)
	var count int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE code LIKE $1`, table)
	if err := r.db.QueryRow(ctx, q, pattern).Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%03d", prefix, year, count+1), nil
}

// Opportunities
func (r *Repository) CreateOpportunity(ctx context.Context, o *Opportunity) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO opportunities (customer_id, title, stage, value, currency, expected_close, assigned_to, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at, updated_at`,
		o.CustomerID, o.Title, o.Stage, o.Value, o.Currency, o.ExpectedClose, o.AssignedTo, o.CreatedBy,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}

func (r *Repository) ListOpportunities(ctx context.Context, stage, assignedTo, userID, role string, limit, offset int) ([]Opportunity, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1
	if stage != "" {
		where += fmt.Sprintf(" AND stage = $%d", n)
		args = append(args, stage)
		n++
	}
	if assignedTo != "" {
		where += fmt.Sprintf(" AND assigned_to = $%d", n)
		args = append(args, assignedTo)
		n++
	}
	if role == "sales_rep" {
		where += fmt.Sprintf(" AND assigned_to = $%d", n)
		args = append(args, userID)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM opportunities "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, customer_id, title, stage, value, currency, expected_close, assigned_to, lost_reason, created_by, created_at, updated_at
		FROM opportunities %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanOpportunities(rows), total, nil
}

func scanOpportunities(rows pgx.Rows) []Opportunity {
	var list []Opportunity
	for rows.Next() {
		var o Opportunity
		_ = rows.Scan(&o.ID, &o.CustomerID, &o.Title, &o.Stage, &o.Value, &o.Currency, &o.ExpectedClose,
			&o.AssignedTo, &o.LostReason, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt)
		list = append(list, o)
	}
	return list
}

func (r *Repository) GetOpportunity(ctx context.Context, id uuid.UUID) (*Opportunity, error) {
	o := &Opportunity{}
	err := r.db.QueryRow(ctx, `
		SELECT id, customer_id, title, stage, value, currency, expected_close, assigned_to, lost_reason, created_by, created_at, updated_at
		FROM opportunities WHERE id=$1`, id,
	).Scan(&o.ID, &o.CustomerID, &o.Title, &o.Stage, &o.Value, &o.Currency, &o.ExpectedClose,
		&o.AssignedTo, &o.LostReason, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return o, err
}

func (r *Repository) UpdateOpportunity(ctx context.Context, o *Opportunity) error {
	_, err := r.db.Exec(ctx, `
		UPDATE opportunities SET title=$1, stage=$2, value=$3, currency=$4, expected_close=$5, assigned_to=$6, lost_reason=$7, updated_at=now()
		WHERE id=$8`, o.Title, o.Stage, o.Value, o.Currency, o.ExpectedClose, o.AssignedTo, o.LostReason, o.ID)
	return err
}

func (r *Repository) InsertStageHistory(ctx context.Context, oppID uuid.UUID, fromStage, toStage string, changedBy *uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO opportunity_stage_history (opportunity_id, from_stage, to_stage, changed_by)
		VALUES ($1, $2, $3, $4)`, oppID, fromStage, toStage, changedBy)
	return err
}

func (r *Repository) DeleteOpportunity(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM opportunities WHERE id=$1`, id)
	return err
}

// Contracts
func (r *Repository) CreateContract(ctx context.Context, c *Contract) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO contracts (code, customer_id, opportunity_id, title, start_date, end_date, service_type, value, currency, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id, created_at, updated_at`,
		c.Code, c.CustomerID, c.OpportunityID, c.Title, c.StartDate, c.EndDate, c.ServiceType, c.Value, c.Currency, c.Status, c.CreatedBy,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *Repository) ListContracts(ctx context.Context, status, customerID string, limit, offset int) ([]Contract, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	if customerID != "" {
		where += fmt.Sprintf(" AND customer_id = $%d", n)
		args = append(args, customerID)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM contracts "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, code, customer_id, opportunity_id, title, start_date, end_date, service_type, value, currency, status, file_url, created_by, created_at, updated_at
		FROM contracts %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Contract
	for rows.Next() {
		var c Contract
		_ = rows.Scan(&c.ID, &c.Code, &c.CustomerID, &c.OpportunityID, &c.Title, &c.StartDate, &c.EndDate,
			&c.ServiceType, &c.Value, &c.Currency, &c.Status, &c.FileURL, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
		list = append(list, c)
	}
	return list, total, nil
}

func (r *Repository) GetContract(ctx context.Context, id uuid.UUID) (*Contract, error) {
	c := &Contract{}
	err := r.db.QueryRow(ctx, `
		SELECT id, code, customer_id, opportunity_id, title, start_date, end_date, service_type, value, currency, status, file_url, created_by, created_at, updated_at
		FROM contracts WHERE id=$1`, id,
	).Scan(&c.ID, &c.Code, &c.CustomerID, &c.OpportunityID, &c.Title, &c.StartDate, &c.EndDate,
		&c.ServiceType, &c.Value, &c.Currency, &c.Status, &c.FileURL, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *Repository) UpdateContract(ctx context.Context, c *Contract) error {
	_, err := r.db.Exec(ctx, `
		UPDATE contracts SET title=$1, start_date=$2, end_date=$3, service_type=$4, value=$5, currency=$6, status=$7, file_url=$8, updated_at=now()
		WHERE id=$9`, c.Title, c.StartDate, c.EndDate, c.ServiceType, c.Value, c.Currency, c.Status, c.FileURL, c.ID)
	return err
}

// Quotations
func (r *Repository) CreateQuotation(ctx context.Context, q *Quotation, itemsJSON []byte) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO quotations (code, customer_id, opp_id, items, total, currency, valid_until, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at`,
		q.Code, q.CustomerID, q.OppID, itemsJSON, q.Total, q.Currency, q.ValidUntil, q.Status, q.CreatedBy,
	).Scan(&q.ID, &q.CreatedAt)
}

func (r *Repository) ListQuotations(ctx context.Context, status, customerID string, limit, offset int) ([]Quotation, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	if customerID != "" {
		where += fmt.Sprintf(" AND customer_id = $%d", n)
		args = append(args, customerID)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM quotations "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, code, customer_id, opp_id, items, total, currency, valid_until, status, created_by, created_at
		FROM quotations %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Quotation
	for rows.Next() {
		var qt Quotation
		var itemsJSON []byte
		_ = rows.Scan(&qt.ID, &qt.Code, &qt.CustomerID, &qt.OppID, &itemsJSON, &qt.Total, &qt.Currency, &qt.ValidUntil, &qt.Status, &qt.CreatedBy, &qt.CreatedAt)
		qt.Items, _ = ItemsFromJSON(itemsJSON)
		list = append(list, qt)
	}
	return list, total, nil
}

func (r *Repository) GetQuotation(ctx context.Context, id uuid.UUID) (*Quotation, error) {
	qt := &Quotation{}
	var itemsJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, code, customer_id, opp_id, items, total, currency, valid_until, status, created_by, created_at
		FROM quotations WHERE id=$1`, id,
	).Scan(&qt.ID, &qt.Code, &qt.CustomerID, &qt.OppID, &itemsJSON, &qt.Total, &qt.Currency, &qt.ValidUntil, &qt.Status, &qt.CreatedBy, &qt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	qt.Items, _ = ItemsFromJSON(itemsJSON)
	return qt, err
}

func (r *Repository) UpdateQuotation(ctx context.Context, qt *Quotation, itemsJSON []byte) error {
	_, err := r.db.Exec(ctx, `
		UPDATE quotations SET items=$1, total=$2, currency=$3, valid_until=$4, status=$5 WHERE id=$6`,
		itemsJSON, qt.Total, qt.Currency, qt.ValidUntil, qt.Status, qt.ID)
	return err
}

func (r *Repository) GetCustomerEmail(ctx context.Context, customerID uuid.UUID) (string, error) {
	var email *string
	err := r.db.QueryRow(ctx, `SELECT email FROM customers WHERE id=$1`, customerID).Scan(&email)
	if err != nil || email == nil {
		return "", err
	}
	return *email, nil
}
