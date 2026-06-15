package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/sales/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContractRepo struct{ db *pgxpool.Pool }

func NewContractRepo(db *pgxpool.Pool) *ContractRepo { return &ContractRepo{db: db} }

func (r *ContractRepo) NextCode(ctx context.Context, prefix, table string) (string, error) {
	year := time.Now().Year()
	pattern := fmt.Sprintf("%s-%d-%%", prefix, year)
	var count int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE code LIKE $1`, table)
	if err := r.db.QueryRow(ctx, q, pattern).Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%03d", prefix, year, count+1), nil
}

func (r *ContractRepo) Create(ctx context.Context, c *domain.Contract) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO contracts (code, customer_id, opportunity_id, title, start_date, end_date, service_type, value, currency, status, payment_terms, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id, created_at, updated_at`,
		c.Code, c.CustomerID, c.OpportunityID, c.Title, c.StartDate, c.EndDate, c.ServiceType, c.Value, c.Currency, c.Status, c.PaymentTerms, c.CreatedBy,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *ContractRepo) List(ctx context.Context, status, customerID string, limit, offset int) ([]domain.Contract, int, error) {
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
	q := fmt.Sprintf(`SELECT id, code, customer_id, opportunity_id, title, start_date, end_date, service_type, value, currency, status, payment_terms, file_url, created_by, created_at, updated_at
		FROM contracts %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []domain.Contract
	for rows.Next() {
		var c domain.Contract
		_ = rows.Scan(&c.ID, &c.Code, &c.CustomerID, &c.OpportunityID, &c.Title, &c.StartDate, &c.EndDate,
			&c.ServiceType, &c.Value, &c.Currency, &c.Status, &c.PaymentTerms, &c.FileURL, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
		list = append(list, c)
	}
	return list, total, nil
}

func (r *ContractRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Contract, error) {
	c := &domain.Contract{}
	err := r.db.QueryRow(ctx, `
		SELECT id, code, customer_id, opportunity_id, title, start_date, end_date, service_type, value, currency, status, payment_terms, file_url, created_by, created_at, updated_at
		FROM contracts WHERE id=$1`, id,
	).Scan(&c.ID, &c.Code, &c.CustomerID, &c.OpportunityID, &c.Title, &c.StartDate, &c.EndDate,
		&c.ServiceType, &c.Value, &c.Currency, &c.Status, &c.PaymentTerms, &c.FileURL, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, applicationErrNotFound
	}
	return c, err
}

func (r *ContractRepo) Update(ctx context.Context, c *domain.Contract) error {
	_, err := r.db.Exec(ctx, `
		UPDATE contracts SET title=$1, start_date=$2, end_date=$3, service_type=$4, value=$5, currency=$6, status=$7, payment_terms=$8, file_url=$9, updated_at=now()
		WHERE id=$10`, c.Title, c.StartDate, c.EndDate, c.ServiceType, c.Value, c.Currency, c.Status, c.PaymentTerms, c.FileURL, c.ID)
	return err
}

type QuotationRepo struct{ db *pgxpool.Pool }

func NewQuotationRepo(db *pgxpool.Pool) *QuotationRepo { return &QuotationRepo{db: db} }

func (r *QuotationRepo) NextCode(ctx context.Context, prefix, table string) (string, error) {
	return (&ContractRepo{db: r.db}).NextCode(ctx, prefix, table)
}

func (r *QuotationRepo) Create(ctx context.Context, q *domain.Quotation, itemsJSON []byte) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO quotations (code, customer_id, opp_id, items, total, currency, valid_until, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at`,
		q.Code, q.CustomerID, q.OppID, itemsJSON, q.Total, q.Currency, q.ValidUntil, q.Status, q.CreatedBy,
	).Scan(&q.ID, &q.CreatedAt)
}

func (r *QuotationRepo) List(ctx context.Context, status, customerID string, limit, offset int) ([]domain.Quotation, int, error) {
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
	var list []domain.Quotation
	for rows.Next() {
		var qt domain.Quotation
		var itemsJSON []byte
		_ = rows.Scan(&qt.ID, &qt.Code, &qt.CustomerID, &qt.OppID, &itemsJSON, &qt.Total, &qt.Currency, &qt.ValidUntil, &qt.Status, &qt.CreatedBy, &qt.CreatedAt)
		_ = json.Unmarshal(itemsJSON, &qt.Items)
		list = append(list, qt)
	}
	return list, total, nil
}

func (r *QuotationRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Quotation, error) {
	qt := &domain.Quotation{}
	var itemsJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, code, customer_id, opp_id, items, total, currency, valid_until, status, created_by, created_at
		FROM quotations WHERE id=$1`, id,
	).Scan(&qt.ID, &qt.Code, &qt.CustomerID, &qt.OppID, &itemsJSON, &qt.Total, &qt.Currency, &qt.ValidUntil, &qt.Status, &qt.CreatedBy, &qt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, applicationErrNotFound
	}
	_ = json.Unmarshal(itemsJSON, &qt.Items)
	return qt, err
}

func (r *QuotationRepo) Update(ctx context.Context, qt *domain.Quotation, itemsJSON []byte) error {
	_, err := r.db.Exec(ctx, `
		UPDATE quotations SET items=$1, total=$2, currency=$3, valid_until=$4, status=$5 WHERE id=$6`,
		itemsJSON, qt.Total, qt.Currency, qt.ValidUntil, qt.Status, qt.ID)
	return err
}

func (r *QuotationRepo) GetCustomerEmail(ctx context.Context, customerID uuid.UUID) (string, error) {
	var email *string
	err := r.db.QueryRow(ctx, `SELECT email FROM customers WHERE id=$1`, customerID).Scan(&email)
	if err != nil || email == nil {
		return "", err
	}
	return *email, nil
}
