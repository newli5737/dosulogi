package crm

import (
	"context"
	"errors"
	"fmt"

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

func (r *Repository) NextCustomerCode(ctx context.Context) (string, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM customers`).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("KH-%04d", count+1), nil
}

func (r *Repository) CreateCustomer(ctx context.Context, c *Customer) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO customers (code, name, type, email, phone, address, province, segment, tier, assigned_to, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at, updated_at`,
		c.Code, c.Name, c.Type, c.Email, c.Phone, c.Address, c.Province, c.Segment, c.Tier, c.AssignedTo, c.CreatedBy,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *Repository) buildCustomerFilter(f CustomerFilter) (string, []interface{}) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1

	if f.Segment != "" {
		where += fmt.Sprintf(" AND segment = $%d", n)
		args = append(args, f.Segment)
		n++
	}
	if f.Tier != "" {
		where += fmt.Sprintf(" AND tier = $%d", n)
		args = append(args, f.Tier)
		n++
	}
	if f.AssignedTo != "" {
		where += fmt.Sprintf(" AND assigned_to = $%d", n)
		args = append(args, f.AssignedTo)
		n++
	}
	if f.Query != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR email ILIKE $%d)", n, n, n)
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if f.Role == "sales_rep" && f.UserID != "" {
		where += fmt.Sprintf(" AND assigned_to = $%d", n)
		args = append(args, f.UserID)
		n++
	}
	return where, args
}

func (r *Repository) ListCustomers(ctx context.Context, f CustomerFilter) ([]Customer, int, error) {
	where, args := r.buildCustomerFilter(f)
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM customers "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT id, code, name, type, email, phone, address, province, segment, tier,
		assigned_to, last_contact_at, created_by, created_at, updated_at
		FROM customers %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanCustomers(rows), total, nil
}

func scanCustomers(rows pgx.Rows) []Customer {
	var list []Customer
	for rows.Next() {
		var c Customer
		_ = rows.Scan(&c.ID, &c.Code, &c.Name, &c.Type, &c.Email, &c.Phone, &c.Address, &c.Province,
			&c.Segment, &c.Tier, &c.AssignedTo, &c.LastContactAt, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
		list = append(list, c)
	}
	return list
}

func (r *Repository) GetCustomer(ctx context.Context, id uuid.UUID) (*Customer, error) {
	c := &Customer{}
	err := r.db.QueryRow(ctx, `
		SELECT id, code, name, type, email, phone, address, province, segment, tier,
		assigned_to, last_contact_at, created_by, created_at, updated_at
		FROM customers WHERE id=$1`, id,
	).Scan(&c.ID, &c.Code, &c.Name, &c.Type, &c.Email, &c.Phone, &c.Address, &c.Province,
		&c.Segment, &c.Tier, &c.AssignedTo, &c.LastContactAt, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *Repository) UpdateCustomer(ctx context.Context, c *Customer) error {
	_, err := r.db.Exec(ctx, `
		UPDATE customers SET name=$1, type=$2, email=$3, phone=$4, address=$5, province=$6,
		segment=$7, tier=$8, assigned_to=$9, updated_at=now() WHERE id=$10`,
		c.Name, c.Type, c.Email, c.Phone, c.Address, c.Province, c.Segment, c.Tier, c.AssignedTo, c.ID)
	return err
}

func (r *Repository) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM customers WHERE id=$1`, id)
	return err
}

func (r *Repository) ListContacts(ctx context.Context, customerID uuid.UUID) ([]Contact, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, customer_id, name, role, phone, email, is_primary, created_at
		FROM contacts WHERE customer_id=$1 ORDER BY is_primary DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Contact
	for rows.Next() {
		var ct Contact
		_ = rows.Scan(&ct.ID, &ct.CustomerID, &ct.Name, &ct.Role, &ct.Phone, &ct.Email, &ct.IsPrimary, &ct.CreatedAt)
		list = append(list, ct)
	}
	return list, nil
}

func (r *Repository) CreateContact(ctx context.Context, ct *Contact) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO contacts (customer_id, name, role, phone, email, is_primary)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at`,
		ct.CustomerID, ct.Name, ct.Role, ct.Phone, ct.Email, ct.IsPrimary,
	).Scan(&ct.ID, &ct.CreatedAt)
}

func (r *Repository) UpdateContact(ctx context.Context, ct *Contact) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE contacts SET name=$1, role=$2, phone=$3, email=$4, is_primary=$5
		WHERE id=$6 AND customer_id=$7`,
		ct.Name, ct.Role, ct.Phone, ct.Email, ct.IsPrimary, ct.ID, ct.CustomerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteContact(ctx context.Context, customerID, contactID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM contacts WHERE id=$1 AND customer_id=$2`, contactID, customerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListInteractions(ctx context.Context, customerID uuid.UUID) ([]Interaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, customer_id, channel, direction, summary, created_by, created_at
		FROM interactions WHERE customer_id=$1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Interaction
	for rows.Next() {
		var i Interaction
		_ = rows.Scan(&i.ID, &i.CustomerID, &i.Channel, &i.Direction, &i.Summary, &i.CreatedBy, &i.CreatedAt)
		list = append(list, i)
	}
	return list, nil
}

func (r *Repository) CreateInteraction(ctx context.Context, i *Interaction) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO interactions (customer_id, channel, direction, summary, created_by)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		i.CustomerID, i.Channel, i.Direction, i.Summary, i.CreatedBy,
	).Scan(&i.ID, &i.CreatedAt)
}

func (r *Repository) TouchLastContact(ctx context.Context, customerID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE customers SET last_contact_at=now(), updated_at=now() WHERE id=$1`, customerID)
	return err
}
