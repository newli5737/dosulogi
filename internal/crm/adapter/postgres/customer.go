package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/dosu-logi/logistics-erp/internal/crm/domain"
	"github.com/dosu-logi/logistics-erp/internal/platform/codegen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepo struct{ db *pgxpool.Pool }

func NewCustomerRepo(db *pgxpool.Pool) *CustomerRepo { return &CustomerRepo{db: db} }

func (r *CustomerRepo) List(ctx context.Context, f domain.CustomerFilter) ([]domain.Customer, int, error) {
	where := "WHERE c.is_active = true"
	args := []any{}
	n := 1
	if f.Query != "" {
		where += fmt.Sprintf(" AND (c.name ILIKE $%d OR c.code ILIKE $%d OR c.email ILIKE $%d OR c.phone ILIKE $%d)", n, n, n, n)
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if f.Type != "" {
		where += fmt.Sprintf(" AND c.type = $%d", n)
		args = append(args, f.Type)
		n++
	}
	if f.Segment != "" {
		where += fmt.Sprintf(" AND c.segment = $%d", n)
		args = append(args, f.Segment)
		n++
	}
	if f.Tier != "" {
		where += fmt.Sprintf(" AND c.tier = $%d", n)
		args = append(args, f.Tier)
		n++
	}
	if f.AssignedTo != "" {
		where += fmt.Sprintf(" AND c.assigned_to = $%d", n)
		args = append(args, f.AssignedTo)
		n++
	}
	if f.Role == "sales_rep" && f.UserID != "" {
		where += fmt.Sprintf(" AND c.assigned_to = $%d", n)
		args = append(args, f.UserID)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM customers c "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT c.id, c.code, c.name, c.type, c.email, c.phone, c.address, c.province, c.tax_code,
		c.segment, c.tier, c.assigned_to, u.id, u.full_name, c.last_contact_at, c.is_active, c.created_at
		FROM customers c LEFT JOIN users u ON u.id = c.assigned_to %s
		ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []domain.Customer
	for rows.Next() {
		var c domain.Customer
		var uid *uuid.UUID
		var uname *string
		_ = rows.Scan(&c.ID, &c.Code, &c.Name, &c.Type, &c.Email, &c.Phone, &c.Address, &c.Province, &c.TaxCode,
			&c.Segment, &c.Tier, &c.AssignedTo, &uid, &uname, &c.LastContactAt, &c.IsActive, &c.CreatedAt)
		if uid != nil && uname != nil {
			c.AssignedUser = &domain.UserBrief{ID: *uid, FullName: *uname}
		}
		list = append(list, c)
	}
	if list == nil {
		list = []domain.Customer{}
	}
	return list, total, nil
}

func (r *CustomerRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	var c domain.Customer
	var uid *uuid.UUID
	var uname *string
	err := r.db.QueryRow(ctx, `SELECT c.id, c.code, c.name, c.type, c.email, c.phone, c.address, c.province, c.tax_code,
		c.segment, c.tier, c.assigned_to, u.id, u.full_name, c.last_contact_at, c.is_active, c.created_at
		FROM customers c LEFT JOIN users u ON u.id = c.assigned_to WHERE c.id=$1 AND c.is_active=true`, id).
		Scan(&c.ID, &c.Code, &c.Name, &c.Type, &c.Email, &c.Phone, &c.Address, &c.Province, &c.TaxCode,
			&c.Segment, &c.Tier, &c.AssignedTo, &uid, &uname, &c.LastContactAt, &c.IsActive, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	if uid != nil && uname != nil {
		c.AssignedUser = &domain.UserBrief{ID: *uid, FullName: *uname}
	}
	return &c, err
}

func (r *CustomerRepo) Create(ctx context.Context, c *domain.Customer) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO customers (code,name,type,email,phone,address,province,tax_code,segment,tier,assigned_to)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id, created_at`,
		c.Code, c.Name, c.Type, c.Email, c.Phone, c.Address, c.Province, c.TaxCode, c.Segment, c.Tier, c.AssignedTo,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *CustomerRepo) Update(ctx context.Context, c *domain.Customer) error {
	_, err := r.db.Exec(ctx, `
		UPDATE customers SET name=$1,type=$2,email=$3,phone=$4,address=$5,province=$6,tax_code=$7,
		segment=$8,tier=$9,assigned_to=$10, updated_at=now() WHERE id=$11`,
		c.Name, c.Type, c.Email, c.Phone, c.Address, c.Province, c.TaxCode, c.Segment, c.Tier, c.AssignedTo, c.ID)
	return err
}

func (r *CustomerRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE customers SET is_active=false, updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *CustomerRepo) EmailExists(ctx context.Context, email string, excludeID *uuid.UUID) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(SELECT 1 FROM customers WHERE email=$1 AND is_active=true`
	args := []any{email}
	if excludeID != nil {
		q += ` AND id<>$2`
		args = append(args, *excludeID)
	}
	q += `)`
	return exists, r.db.QueryRow(ctx, q, args...).Scan(&exists)
}

func (r *CustomerRepo) NextCode(ctx context.Context) (string, error) {
	return codegen.Next(ctx, r.db, "customers", "KH", false)
}
