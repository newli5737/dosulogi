package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/crm/domain"
	"github.com/dosu-logi/logistics-erp/internal/platform/codegen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepo struct{ db *pgxpool.Pool }

func NewTicketRepo(db *pgxpool.Pool) *TicketRepo { return &TicketRepo{db: db} }

func (r *TicketRepo) List(ctx context.Context, f domain.TicketFilter) ([]domain.Ticket, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	n := 1
	if f.Status != "" {
		where += fmt.Sprintf(" AND t.status = $%d", n)
		args = append(args, f.Status)
		n++
	}
	if f.Priority != "" {
		where += fmt.Sprintf(" AND t.priority = $%d", n)
		args = append(args, f.Priority)
		n++
	}
	if f.CustomerID != "" {
		where += fmt.Sprintf(" AND t.customer_id = $%d", n)
		args = append(args, f.CustomerID)
		n++
	}
	if f.Overdue {
		where += " AND t.sla_deadline < now() AND t.status NOT IN ('resolved','closed')"
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM tickets t "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT t.id,t.code,t.customer_id,t.title,t.description,t.priority,t.status,t.category,t.assigned_to,t.sla_deadline,t.created_at,
		c.id,c.name,c.code,u.id,u.full_name FROM tickets t
		JOIN customers c ON c.id=t.customer_id LEFT JOIN users u ON u.id=t.assigned_to %s ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	now := time.Now()
	var list []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		var cid uuid.UUID
		var cname, ccode string
		var uid *uuid.UUID
		var uname *string
		_ = rows.Scan(&t.ID, &t.Code, &t.CustomerID, &t.Title, &t.Description, &t.Priority, &t.Status, &t.Category, &t.AssignedTo, &t.SLADeadline, &t.CreatedAt, &cid, &cname, &ccode, &uid, &uname)
		t.Customer = &struct {
			ID   uuid.UUID `json:"id"`
			Name string    `json:"name"`
			Code string    `json:"code"`
		}{ID: cid, Name: cname, Code: ccode}
		if uid != nil && uname != nil {
			t.AssignedUser = &domain.UserBrief{ID: *uid, FullName: *uname}
		}
		if t.SLADeadline != nil && t.Status != "resolved" && t.Status != "closed" {
			t.IsOverdue = t.SLADeadline.Before(now)
		}
		list = append(list, t)
	}
	if list == nil {
		list = []domain.Ticket{}
	}
	return list, total, nil
}

func (r *TicketRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	var t domain.Ticket
	err := r.db.QueryRow(ctx, `SELECT id,code,customer_id,title,description,priority,status,category,assigned_to,sla_deadline,created_at FROM tickets WHERE id=$1`, id).
		Scan(&t.ID, &t.Code, &t.CustomerID, &t.Title, &t.Description, &t.Priority, &t.Status, &t.Category, &t.AssignedTo, &t.SLADeadline, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	return &t, err
}

func (r *TicketRepo) Create(ctx context.Context, t *domain.Ticket) error {
	return r.db.QueryRow(ctx, `INSERT INTO tickets (code,customer_id,title,description,priority,status,category,assigned_to,sla_deadline) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id,created_at`, t.Code, t.CustomerID, t.Title, t.Description, t.Priority, t.Status, t.Category, t.AssignedTo, t.SLADeadline).Scan(&t.ID, &t.CreatedAt)
}

func (r *TicketRepo) Update(ctx context.Context, t *domain.Ticket) error {
	_, err := r.db.Exec(ctx, `UPDATE tickets SET status=$1,priority=$2,assigned_to=$3,sla_deadline=$4,updated_at=now() WHERE id=$5`, t.Status, t.Priority, t.AssignedTo, t.SLADeadline, t.ID)
	return err
}

func (r *TicketRepo) NextCode(ctx context.Context) (string, error) {
	return codegen.Next(ctx, r.db, "tickets", "TK", false)
}

func (r *TicketRepo) GetCustomerAssignee(ctx context.Context, customerID uuid.UUID) (*uuid.UUID, error) {
	var id *uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT assigned_to FROM customers WHERE id=$1`, customerID).Scan(&id)
	return id, err
}

type TicketCommentRepo struct{ db *pgxpool.Pool }

func NewTicketCommentRepo(db *pgxpool.Pool) *TicketCommentRepo { return &TicketCommentRepo{db: db} }

func (r *TicketCommentRepo) ListByTicket(ctx context.Context, ticketID uuid.UUID) ([]domain.TicketComment, error) {
	rows, err := r.db.Query(ctx, `SELECT tc.id,tc.ticket_id,tc.body,tc.is_internal,tc.created_at,u.id,u.full_name FROM ticket_comments tc LEFT JOIN users u ON u.id=tc.created_by WHERE tc.ticket_id=$1 ORDER BY tc.created_at`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.TicketComment
	for rows.Next() {
		var c domain.TicketComment
		var uid *uuid.UUID
		var uname *string
		_ = rows.Scan(&c.ID, &c.TicketID, &c.Body, &c.IsInternal, &c.CreatedAt, &uid, &uname)
		if uid != nil && uname != nil {
			c.CreatedBy = &domain.UserBrief{ID: *uid, FullName: *uname}
		}
		list = append(list, c)
	}
	if list == nil {
		list = []domain.TicketComment{}
	}
	return list, nil
}

func (r *TicketCommentRepo) Create(ctx context.Context, c *domain.TicketComment, createdBy uuid.UUID) error {
	return r.db.QueryRow(ctx, `INSERT INTO ticket_comments (ticket_id,body,is_internal,created_by) VALUES ($1,$2,$3,$4) RETURNING id,created_at`, c.TicketID, c.Body, c.IsInternal, createdBy).Scan(&c.ID, &c.CreatedAt)
}

type ContactRepo struct{ db *pgxpool.Pool }

func NewContactRepo(db *pgxpool.Pool) *ContactRepo { return &ContactRepo{db: db} }

func (r *ContactRepo) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Contact, error) {
	rows, err := r.db.Query(ctx, `SELECT id, customer_id, name, role, phone, email, is_primary, note FROM contacts WHERE customer_id=$1 ORDER BY is_primary DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Contact
	for rows.Next() {
		var c domain.Contact
		_ = rows.Scan(&c.ID, &c.CustomerID, &c.Name, &c.Role, &c.Phone, &c.Email, &c.IsPrimary, &c.Note)
		list = append(list, c)
	}
	if list == nil {
		list = []domain.Contact{}
	}
	return list, nil
}

func (r *ContactRepo) Create(ctx context.Context, c *domain.Contact) error {
	return r.db.QueryRow(ctx, `INSERT INTO contacts (customer_id,name,role,phone,email,is_primary,note) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, c.CustomerID, c.Name, c.Role, c.Phone, c.Email, c.IsPrimary, c.Note).Scan(&c.ID)
}

func (r *ContactRepo) Update(ctx context.Context, c *domain.Contact) error {
	_, err := r.db.Exec(ctx, `UPDATE contacts SET name=$1,role=$2,phone=$3,email=$4,is_primary=$5,note=$6,updated_at=now() WHERE id=$7 AND customer_id=$8`, c.Name, c.Role, c.Phone, c.Email, c.IsPrimary, c.Note, c.ID, c.CustomerID)
	return err
}

func (r *ContactRepo) Delete(ctx context.Context, customerID, contactID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM contacts WHERE id=$1 AND customer_id=$2`, contactID, customerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ContactRepo) UnsetPrimary(ctx context.Context, customerID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE contacts SET is_primary=false WHERE customer_id=$1`, customerID)
	return err
}

func (r *ContactRepo) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error) {
	var n int
	return n, r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contacts WHERE customer_id=$1`, customerID).Scan(&n)
}
