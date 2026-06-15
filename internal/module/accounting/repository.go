package accounting

import (
	"context"
	"encoding/json"
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

func (r *Repository) NextInvoiceCode(ctx context.Context) (string, error) {
	year := time.Now().Year()
	var count int
	pattern := fmt.Sprintf("INV-%d-%%", year)
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM invoices WHERE code LIKE $1`, pattern).Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%d-%03d", year, count+1), nil
}

func (r *Repository) CreateInvoice(ctx context.Context, inv *Invoice) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO invoices (code, customer_id, contract_id, shipment_id, items, subtotal, tax_rate, tax_amount, total, currency, status, due_date, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id, created_at, updated_at`,
		inv.Code, inv.CustomerID, inv.ContractID, inv.ShipmentID, inv.Items, inv.Subtotal, inv.TaxRate, inv.TaxAmount, inv.Total, inv.Currency, inv.Status, inv.DueDate, inv.CreatedBy,
	).Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)
}

func (r *Repository) ListInvoices(ctx context.Context, status, customerID, from, to string, limit, offset int) ([]Invoice, int, error) {
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
	if from != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", n)
		args = append(args, from)
		n++
	}
	if to != "" {
		where += fmt.Sprintf(" AND created_at <= $%d", n)
		args = append(args, to)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM invoices "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, code, customer_id, contract_id, shipment_id, items, subtotal, tax_rate, tax_amount, total, currency, status, due_date, paid_at, file_url, created_by, created_at, updated_at
		FROM invoices %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Invoice
	for rows.Next() {
		var inv Invoice
		_ = rows.Scan(&inv.ID, &inv.Code, &inv.CustomerID, &inv.ContractID, &inv.ShipmentID, &inv.Items, &inv.Subtotal, &inv.TaxRate, &inv.TaxAmount, &inv.Total, &inv.Currency, &inv.Status, &inv.DueDate, &inv.PaidAt, &inv.FileURL, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt)
		list = append(list, inv)
	}
	return list, total, nil
}

func (r *Repository) GetInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	inv := &Invoice{}
	err := r.db.QueryRow(ctx, `
		SELECT id, code, customer_id, contract_id, shipment_id, items, subtotal, tax_rate, tax_amount, total, currency, status, due_date, paid_at, file_url, created_by, created_at, updated_at
		FROM invoices WHERE id=$1`, id,
	).Scan(&inv.ID, &inv.Code, &inv.CustomerID, &inv.ContractID, &inv.ShipmentID, &inv.Items, &inv.Subtotal, &inv.TaxRate, &inv.TaxAmount, &inv.Total, &inv.Currency, &inv.Status, &inv.DueDate, &inv.PaidAt, &inv.FileURL, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return inv, err
}

func (r *Repository) GetInvoiceByCode(ctx context.Context, code string) (*Invoice, error) {
	inv := &Invoice{}
	err := r.db.QueryRow(ctx, `
		SELECT id, code, customer_id, contract_id, shipment_id, items, subtotal, tax_rate, tax_amount, total, currency, status, due_date, paid_at, file_url, created_by, created_at, updated_at
		FROM invoices WHERE code=$1`, code,
	).Scan(&inv.ID, &inv.Code, &inv.CustomerID, &inv.ContractID, &inv.ShipmentID, &inv.Items, &inv.Subtotal, &inv.TaxRate, &inv.TaxAmount, &inv.Total, &inv.Currency, &inv.Status, &inv.DueDate, &inv.PaidAt, &inv.FileURL, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return inv, err
}

func (r *Repository) UpdateInvoice(ctx context.Context, inv *Invoice) error {
	_, err := r.db.Exec(ctx, `
		UPDATE invoices SET items=$1, subtotal=$2, tax_rate=$3, tax_amount=$4, total=$5, currency=$6, status=$7, due_date=$8, paid_at=$9, file_url=$10, updated_at=now()
		WHERE id=$11`, inv.Items, inv.Subtotal, inv.TaxRate, inv.TaxAmount, inv.Total, inv.Currency, inv.Status, inv.DueDate, inv.PaidAt, inv.FileURL, inv.ID)
	return err
}

func (r *Repository) CreatePayment(ctx context.Context, p *Payment) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO payments (invoice_id, amount, method, reference_code, sepay_txn_id, matched_auto, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		p.InvoiceID, p.Amount, p.Method, p.ReferenceCode, p.SePayTxnID, p.MatchedAuto, p.Note,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *Repository) ListPayments(ctx context.Context, limit, offset int) ([]Payment, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM payments`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, invoice_id, amount, method, reference_code, sepay_txn_id, matched_auto, note, created_at
		FROM payments ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Payment
	for rows.Next() {
		var p Payment
		_ = rows.Scan(&p.ID, &p.InvoiceID, &p.Amount, &p.Method, &p.ReferenceCode, &p.SePayTxnID, &p.MatchedAuto, &p.Note, &p.CreatedAt)
		list = append(list, p)
	}
	return list, total, nil
}

func (r *Repository) SaveUnmatchedPayment(ctx context.Context, sepayTxnID string, amount float64, refCode string, raw []byte) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO unmatched_payments (sepay_txn_id, amount, reference_code, raw_payload)
		VALUES ($1,$2,$3,$4)`, sepayTxnID, amount, refCode, raw)
	return err
}

func (r *Repository) RevenueReport(ctx context.Context, from, to, groupBy string) ([]RevenueReport, error) {
	var q string
	switch groupBy {
	case "customer":
		q = `SELECT COALESCE(c.name, 'Unknown'), COALESCE(SUM(i.total), 0)
			FROM invoices i LEFT JOIN customers c ON c.id = i.customer_id
			WHERE i.status = 'paid' AND i.created_at >= $1::timestamptz AND i.created_at <= $2::timestamptz
			GROUP BY c.name ORDER BY 2 DESC`
	default:
		q = `SELECT to_char(date_trunc('month', created_at), 'YYYY-MM'), COALESCE(SUM(total), 0)
			FROM invoices WHERE status = 'paid' AND created_at >= $1::timestamptz AND created_at <= $2::timestamptz
			GROUP BY 1 ORDER BY 1`
	}
	if from == "" {
		from = time.Now().AddDate(-1, 0, 0).Format(time.RFC3339)
	}
	if to == "" {
		to = time.Now().Format(time.RFC3339)
	}
	rows, err := r.db.Query(ctx, q, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []RevenueReport
	for rows.Next() {
		var rpt RevenueReport
		_ = rows.Scan(&rpt.Label, &rpt.Amount)
		list = append(list, rpt)
	}
	return list, nil
}

func (r *Repository) ARReport(ctx context.Context) ([]ARReport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.name, COALESCE(SUM(i.total), 0), COUNT(i.id)
		FROM invoices i JOIN customers c ON c.id = i.customer_id
		WHERE i.status IN ('sent', 'overdue')
		GROUP BY c.id, c.name ORDER BY 3 DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ARReport
	for rows.Next() {
		var r ARReport
		_ = rows.Scan(&r.CustomerID, &r.CustomerName, &r.TotalDue, &r.InvoiceCount)
		list = append(list, r)
	}
	return list, nil
}

func (r *Repository) GetCustomerName(ctx context.Context, id uuid.UUID) (string, error) {
	var name string
	err := r.db.QueryRow(ctx, `SELECT name FROM customers WHERE id=$1`, id).Scan(&name)
	return name, err
}

func (r *Repository) GetCustomerEmail(ctx context.Context, id uuid.UUID) (string, error) {
	var email *string
	err := r.db.QueryRow(ctx, `SELECT email FROM customers WHERE id=$1`, id).Scan(&email)
	if err != nil || email == nil {
		return "", err
	}
	return *email, nil
}

func (r *Repository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]Invoice, error) {
	list, _, err := r.ListInvoices(ctx, "", customerID.String(), "", "", 100, 0)
	return list, err
}

func ParseItems(data json.RawMessage) ([]LineItem, error) {
	var items []LineItem
	err := json.Unmarshal(data, &items)
	return items, err
}
